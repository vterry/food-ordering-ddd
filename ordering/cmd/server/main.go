package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vterry/food-project/ordering/internal/adapters/api"
	apigen "github.com/vterry/food-project/ordering/internal/adapters/api/generated"
	"github.com/vterry/food-project/ordering/internal/adapters/db"
	"github.com/vterry/food-project/ordering/internal/adapters/db/repository"
	"github.com/vterry/food-project/ordering/internal/adapters/external"
	"github.com/vterry/food-project/ordering/internal/adapters/messaging"
	"github.com/vterry/food-project/ordering/internal/core/services"
)

const (
	defaultPort         = "8083" // Different port to avoid conflict
	defaultCustomerGRPC = "localhost:50051"
	defaultDSN          = "root:root@tcp(localhost:3306)/food_project?parseTime=true"
	defaultMQ           = "amqp://guest:guest@localhost:5672/"
)

func main() {
	port := getEnv("PORT", defaultPort)
	customerGRPCAddr := getEnv("CUSTOMER_GRPC_ADDR", defaultCustomerGRPC)
	dsn := getEnv("MYSQL_DSN", defaultDSN)
	mqURL := getEnv("RABBITMQ_URL", defaultMQ)

	// 1. Database
	database, err := db.NewMySQLConnection(dsn)
	if err != nil {
		slog.Error("failed to connect to mysql", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	// 2. Messaging
	commandPublisher, err := messaging.NewRabbitMQCommandPublisher(mqURL)
	if err != nil {
		slog.Error("failed to connect to rabbitmq publisher", "error", err)
		os.Exit(1)
	}
	defer commandPublisher.Close()

	// 3. External gRPC Client
	customerClient, err := external.NewCustomerGRPCClient(customerGRPCAddr)
	if err != nil {
		slog.Error("failed to connect to customer gRPC", "error", err)
		os.Exit(1)
	}

	// 4. Dependency Injection
	orderRepo := repository.NewSQLOrderRepository(database)
	sagaRepo := repository.NewSQLSagaRepository(database)

	orderSvc := services.NewOrderService(orderRepo, sagaRepo, commandPublisher, customerClient)
	sagaCoordinator := services.NewOrderSagaCoordinator(orderRepo, sagaRepo, commandPublisher)

	// 5. Consumers
	eventConsumer, err := messaging.NewRabbitMQEventConsumer(mqURL, sagaCoordinator)
	if err != nil {
		slog.Error("failed to initialize event consumer", "error", err)
	} else {
		defer eventConsumer.Close()
		if err := eventConsumer.Start(context.Background()); err != nil {
			slog.Error("failed to start event consumer", "error", err)
		}
	}

	// Mock Delivery Consumer
	if getEnv("MOCK_DELIVERY", "false") == "true" {
		mockDelivery, err := messaging.NewMockDeliveryConsumer(mqURL)
		if err == nil {
			defer mockDelivery.Close()
			_ = mockDelivery.Start(context.Background())
			slog.Info("Mock Delivery Consumer started")
		}
	}

	// 6. HTTP Server
	handler := api.NewOrderHandler(orderSvc)
	strictHandler := apigen.NewStrictHandler(handler, nil)
	e := api.NewEcho()
	apigen.RegisterHandlersWithBaseURL(e, strictHandler, "/api/v1")

	go func() {
		slog.Info("Starting Ordering HTTP server", "port", port)
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("shuting down the server: %v", err)
		}
	}()

	// 7. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("error shutting down http", "error", err)
	}

	slog.Info("Ordering Server gracefully stopped")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
