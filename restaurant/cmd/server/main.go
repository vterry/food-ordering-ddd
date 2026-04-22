package main
import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	pb "github.com/vterry/food-project/restaurant/api/proto"
	"github.com/vterry/food-project/restaurant/internal/adapters/api"
	apigen "github.com/vterry/food-project/restaurant/internal/adapters/api/generated"
	customgrpc "github.com/vterry/food-project/restaurant/internal/adapters/api/grpc"
	"github.com/vterry/food-project/restaurant/internal/adapters/db"
	"github.com/vterry/food-project/restaurant/internal/adapters/db/repository"
	"github.com/vterry/food-project/restaurant/internal/adapters/messaging"
	"github.com/vterry/food-project/restaurant/internal/core/services"
	)

const (
	defaultHTTPPort = "8081"
	defaultGRPCPort = "50052"
	defaultDSN      = "root:root@tcp(localhost:3306)/food_project?parseTime=true"
	defaultMQ       = "amqp://guest:guest@localhost:5672/"
)

func main() {
	httpPort := getEnv("HTTP_PORT", defaultHTTPPort)
	grpcPort := getEnv("GRPC_PORT", defaultGRPCPort)
	dsn := getEnv("MYSQL_DSN", defaultDSN)
	mqURL := getEnv("RABBITMQ_URL", defaultMQ)

	// DB Init
	mysqlDB, err := db.NewMySQLConnection(dsn)
	if err != nil {
		slog.Error("failed to connect to mysql", "error", err)
		os.Exit(1)
	}
	defer mysqlDB.Close()

	// Repositories
	restRepo := repository.NewSQLRestaurantRepository(mysqlDB)
	menuRepo := repository.NewSQLMenuRepository(mysqlDB)
	ticketRepo := repository.NewSQLTicketRepository(mysqlDB)

	// Services
	restSvc := services.NewRestaurantService(restRepo, menuRepo)
	ticketSvc := services.NewTicketService(ticketRepo)
	querySvc := services.NewRestaurantQueryService(restRepo, menuRepo)

	// Messaging
	consumer, err := messaging.NewRabbitMQConsumer(mqURL, ticketSvc)
	if err != nil {
		slog.Warn("failed to connect to rabbitmq consumer", "error", err)
	} else {
		defer consumer.Close()
		if err := consumer.Start(context.Background()); err != nil {
			slog.Warn("failed to start rabbitmq consumer", "error", err)
		}
	}

	// Servers
	// 1. HTTP Server (Public Catalog & Admin)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	e := api.NewEcho()

	restaurantHandler := api.NewRestaurantHandler(restSvc, ticketSvc, querySvc)
	strictHandler := apigen.NewStrictHandler(restaurantHandler, nil)
	apigen.RegisterHandlersWithBaseURL(e, strictHandler, "/api/v1")

	go func() {
		slog.Info("Starting HTTP server", "port", httpPort)
		if err := e.Start(":" + httpPort); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start http server", "error", err)
		}
	}()

	// 2. gRPC Server (Internal Service-to-Service)
	grpcServer := grpc.NewServer()
	catalogGRPC := customgrpc.NewCatalogGRPCServer(querySvc)
	pb.RegisterCatalogServiceServer(grpcServer, catalogGRPC)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		slog.Error("failed to listen for grpc", "error", err)
		os.Exit(1)
	}

	go func() {
		slog.Info("Starting gRPC server", "port", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("failed to serve grpc", "error", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("error shutting down http", "error", err)
	}

	grpcServer.GracefulStop()
	slog.Info("Servers stopped")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
