package server

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"

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
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type CatalogServer struct {
	cfg          *config.Config
	db           *sql.DB
	httpServer   *http.Server
	grpcServer   *grpc.Server
	healthCheck  *health.Server
	rabbitCon    *amqp.Connection
	pubChan      *amqp.Channel
	logger       *slog.Logger
	workerCtx    context.Context
	workerCancel context.CancelFunc
	NotifyReady  chan struct{}
	readyOnce    sync.Once
}

func NewCatalogServer(cfg *config.Config, db *sql.DB, rabbitCon *amqp.Connection, pubChan *amqp.Channel, logger *slog.Logger) *CatalogServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &CatalogServer{
		cfg:          cfg,
		db:           db,
		rabbitCon:    rabbitCon,
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

	publisher, err := messaging.NewRabbitMQPublisher(c.pubChan, c.cfg.OutboxCfg.ExchangeName, c.logger)
	if err != nil {
		return fmt.Errorf("failed to create rabbit publisher: %w", err)
	}

	relay := app.NewOutboxProcessor(outboxRepo, publisher, unitOfWork, c.cfg.OutboxCfg.PollingInterval, c.cfg.OutboxCfg.BatchSize, c.cfg.OutboxCfg.RetryCount, c.logger)
	relay.Start(c.workerCtx)

	mux := http.NewServeMux()
	restMenuHandler.RegisterRoutes(mux)
	restRestaurantHandler.RegisterRoutes(mux)

	c.httpServer = &http.Server{
		Addr:    c.cfg.HttpListener,
		Handler: mux,
	}

	httpListener, err := net.Listen("tcp", c.cfg.HttpListener)
	if err != nil {
		return err
	}

	// Create gRPC server BEFORE starting the goroutine to prevent race on Stop()
	c.grpcServer = grpc.NewServer()
	catalogServer := grpcAdapter.NewCatalogGrpcServer(catalogQueryService)
	pb.RegisterCatalogServiceServer(c.grpcServer, catalogServer)

	go func() {
		grpcListener, err := net.Listen("tcp", c.cfg.GrpcListener)
		if err != nil {
			c.logger.Error("Failed to listen for gRPC", "error", err)
			return
		}

		c.logger.Info("gRPC server running", "addr", c.cfg.GrpcListener)

		c.healthCheck = health.NewServer()
		c.healthCheck.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
		grpc_health_v1.RegisterHealthServer(c.grpcServer, c.healthCheck)

		if err := c.grpcServer.Serve(grpcListener); err != nil {
			c.logger.Error("gRPC server failed", "error", err)
		}
	}()

	c.logger.Info("HTTP server running and listening", "addr", c.cfg.HttpListener)

	c.readyOnce.Do(func() {
		close(c.NotifyReady)
	})

	return c.httpServer.Serve(httpListener)
}

func (c *CatalogServer) Stop(ctx context.Context) error {
	c.logger.Info("Stopping background workers ... ")
	c.workerCancel()

	if c.healthCheck != nil {
		c.logger.Info("Changing gRPC health status to NOT_SERVING...")
		c.healthCheck.SetServingStatus("", *grpc_health_v1.HealthCheckResponse_NOT_SERVING.Enum())
	}

	//Shutting down HTTP Server
	c.logger.Info("Shutting down catalog server...")
	var httpErr error
	if c.httpServer != nil {
		c.logger.Info("Shutting down HTTP server ...")
		httpErr = c.httpServer.Shutdown(ctx)
	}

	// Shutting down gRPC Server
	if c.grpcServer != nil {
		c.logger.Info("Shutting down gRPC server ...")

		stopped := make(chan struct{})

		go func() {
			c.grpcServer.GracefulStop()
			close(stopped)
		}()

		select {
		case <-ctx.Done():
			c.logger.Warn("gRPC GracefulStop timed out, forcing stop...")
			c.grpcServer.Stop()
		case <-stopped:
			c.logger.Info("gRPC server stopped gracefully")
		}

	}

	// Shutting down RabbitMQ Channels and Conns
	c.logger.Info("Closing infra connections ...")
	if c.pubChan != nil {
		c.pubChan.Close()
	}
	if c.rabbitCon != nil {
		c.rabbitCon.Close()
	}

	return httpErr
}
