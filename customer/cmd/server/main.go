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

	pb "github.com/vterry/food-project/customer/api/proto"
	"github.com/vterry/food-project/customer/internal/adapters/api"
	apigen "github.com/vterry/food-project/customer/internal/adapters/api/generated"
	custgrpc "github.com/vterry/food-project/customer/internal/adapters/api/grpc"
	customerDB "github.com/vterry/food-project/customer/internal/adapters/db"
	"github.com/vterry/food-project/customer/internal/adapters/db/repository"
	"github.com/vterry/food-project/customer/internal/adapters/external"
	"github.com/vterry/food-project/customer/internal/adapters/messaging"
	"github.com/vterry/food-project/customer/internal/core/services"
	"google.golang.org/grpc"
)

const (
	defaultPort         = "8080"
	defaultGRPCPort     = "50051"
	defaultRestGRPCPort = "50052"
	defaultDSN          = "root:root@tcp(localhost:3306)/food_project?parseTime=true"
	defaultMQ           = "amqp://guest:guest@localhost:5672/"
)

func main() {
	// 1. Config (Basic env load)
	port := getEnv("PORT", defaultPort)
	grpcPort := getEnv("GRPC_PORT", defaultGRPCPort)
	restGRPCPort := getEnv("RESTAURANT_GRPC_PORT", defaultRestGRPCPort)
	restGRPCAddr := getEnv("RESTAURANT_GRPC_ADDR", "localhost:"+restGRPCPort)
	dsn := getEnv("MYSQL_DSN", defaultDSN)
	mqURL := getEnv("RABBITMQ_URL", defaultMQ)

	// 2. Infra Initialization
	mysqlDB, err := customerDB.NewMySQLConnection(dsn)
	if err != nil {
		slog.Error("failed to connect to mysql", "error", err)
		os.Exit(1)
	}
	defer mysqlDB.Close()

	publisher, err := messaging.NewRabbitMQPublisher(mqURL, "customer-service")
	if err != nil {
		slog.Error("failed to connect to rabbitmq", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()

	// 3. Dependency Injection
	custRepo := repository.NewSQLCustomerRepository(mysqlDB)
	cartRepo := repository.NewSQLCartRepository(mysqlDB)
	catalogClient, err := external.NewRestaurantCatalogClient(restGRPCAddr)
	if err != nil {
		slog.Error("failed to initialize restaurant catalog client", "error", err)
		os.Exit(1)
	}

	custSvc := services.NewCustomerService(custRepo, publisher)
	cartSvc := services.NewCartService(cartRepo, publisher, catalogClient)

	handler := api.NewCustomerHandler(custSvc, cartSvc)

	// 4. Server Setup
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	e := api.NewEcho()

	// Register Handlers
	strictHandler := apigen.NewStrictHandler(handler, nil)
	apigen.RegisterHandlersWithBaseURL(e, strictHandler, "/api/v1")

	// 5. Start Servers with Graceful Shutdown
	// HTTP Server
	go func() {
		slog.Info("Starting HTTP server", "port", port)
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			slog.Error("shutting down the server", "error", err)
		}
	}()

	// gRPC Server
	grpcServer := grpc.NewServer()
	custGRPC := custgrpc.NewCustomerGRPCServer(custSvc)
	pb.RegisterCustomerServiceServer(grpcServer, custGRPC)

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

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("error shutting down http", "error", err)
	}

	grpcServer.GracefulStop()

	slog.Info("Server gracefully stopped")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
