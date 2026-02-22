package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
	grpcAdapter "github.com/vterry/food-ordering/catalog/internal/adapters/input/grpc"
	"github.com/vterry/food-ordering/catalog/internal/adapters/input/grpc/pb"
	"github.com/vterry/food-ordering/catalog/internal/adapters/input/rest"
	"github.com/vterry/food-ordering/catalog/internal/adapters/output/messaging"
	"github.com/vterry/food-ordering/catalog/internal/adapters/output/repository"
	"github.com/vterry/food-ordering/catalog/internal/core/app"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/services"
	"github.com/vterry/food-ordering/catalog/internal/infra/config"
	"google.golang.org/grpc"
)

type CatalogServer struct {
	cfg          *config.Config
	db           *sql.DB
	httpServer   *http.Server
	grpcServer   *grpc.Server
	rabbitCon    *amqp.Connection
	pubChan      *amqp.Channel
	logger       *slog.Logger
	workerCtx    context.Context
	workerCancel context.CancelFunc
	NotifyReady  chan struct{}
}

func NewCatalogServer(cfg *config.Config, db *sql.DB, rabbitCon *amqp.Connection, pubChan *amqp.Channel, logger *slog.Logger) *CatalogServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &CatalogServer{
		cfg:          cfg,
		db:           db,
		pubChan:      pubChan,
		logger:       logger,
		workerCtx:    ctx,
		workerCancel: cancel,
		NotifyReady:  make(chan struct{}),
	}
}

func (c *CatalogServer) Run() error {
	outboxRepo := repository.NewOutboxRepository(c.db)
	menuRepo := repository.NewMenuRepository(c.db, outboxRepo)
	restaurantRepo := repository.NewRestaurantRepository(c.db, outboxRepo)
	catalogQueryRepo := repository.NewCatalogQueryRepository(c.db)
	unitOfWork := repository.NewUnitOfWork(c.db)

	assignMenuService := services.NewMenuAssignmentService()

	menuAppService := app.NewMenuAppService(assignMenuService, unitOfWork, menuRepo, restaurantRepo)
	restaurantAppService := app.NewRestaurantAppService(unitOfWork, restaurantRepo)
	catalogQueryService := app.NewCatalogQueryAppService(catalogQueryRepo)

	restMenuHandler := rest.NewMenuHandler(menuAppService, c.logger)
	restRestaurantHandler := rest.NewRestaurantHandler(restaurantAppService, c.logger)

	publisher, err := messaging.NewRabbitMQPublisher(c.pubChan, c.cfg.OutboxCfg.QueueName, c.logger)
	if err != nil {
		if err != nil {
			return fmt.Errorf("failed to create rabbit publisher: %w", err)
		}
	}

	relay := app.NewOutboxProcessor(outboxRepo, publisher, unitOfWork, c.cfg.OutboxCfg.PollingInterval, c.cfg.OutboxCfg.BatchSize)
	relay.Start(c.workerCtx)

	mux := http.NewServeMux()
	restMenuHandler.RegisterRoutes(mux)
	restRestaurantHandler.RegisterRoutes(mux)

	c.httpServer = &http.Server{
		Addr:    c.cfg.HttpListener,
		Handler: mux,
	}

	listener, err := net.Listen("tcp", c.cfg.HttpListener)
	if err != nil {
		return err
	}

	go func() {
		lis, err := net.Listen("tcp", c.cfg.GrpcListener)
		if err != nil {
			c.logger.Error("Failed to listen to gRPC", "error", err)
			return
		}

		grpcServer := grpc.NewServer()
		catalogServer := grpcAdapter.NewCatalogGrpcServer(catalogQueryService)
		pb.RegisterCatalogServiceServer(grpcServer, catalogServer)
		c.grpcServer = grpcServer

		c.logger.Info("gRPC server running", "addr", c.cfg.GrpcListener)
		if err := grpcServer.Serve(lis); err != nil {
			c.logger.Error("gRPC server failed", "error", err)
		}
	}()

	c.logger.Info("HTTP server running and listening", "addr", c.cfg.HttpListener)

	close(c.NotifyReady)

	return c.httpServer.Serve(listener)
}

func (c *CatalogServer) Stop(ctx context.Context) error {

	c.workerCancel()

	if c.pubChan != nil {
		c.pubChan.Close()
	}

	if c.rabbitCon != nil {
		c.rabbitCon.Close()
	}

	if c.grpcServer != nil {
		c.grpcServer.GracefulStop()
	}

	if c.httpServer != nil {
		return c.httpServer.Shutdown(ctx)
	}
	return nil
}
