package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	commonMessaging "github.com/vterry/food-project/common/pkg/messaging"
	"github.com/vterry/food-project/common/pkg/outbox"
	"github.com/vterry/food-project/delivery/internal/adapters/api"
	"github.com/vterry/food-project/delivery/internal/adapters/db"
	"github.com/vterry/food-project/delivery/internal/adapters/db/repository"
	"github.com/vterry/food-project/delivery/internal/adapters/messaging"
	"github.com/vterry/food-project/delivery/internal/core/services"
)

const (
	defaultPort = "8084"
	defaultDSN  = "root:root@tcp(localhost:3306)/food_project?parseTime=true"
	defaultMQ   = "amqp://guest:guest@localhost:5672/"
)

func main() {
	// 1. Config
	port := getEnv("PORT", defaultPort)
	dsn := getEnv("MYSQL_DSN", defaultDSN)
	mqURL := getEnv("RABBITMQ_URL", defaultMQ)

	// 2. Logger Setup
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 3. Infra Initialization
	mysqlDB, err := db.NewMySQLConnection(dsn)
	if err != nil {
		slog.Error("failed to connect to mysql", "error", err)
		os.Exit(1)
	}
	defer mysqlDB.Close()

	publisher, err := messaging.NewRabbitMQEventPublisher(mqURL, "delivery-service")
	if err != nil {
		slog.Error("failed to connect to rabbitmq publisher", "error", err)
		os.Exit(1)
	}
	defer publisher.Close()

	// 4. Dependency Injection
	deliveryRepo := repository.NewMySQLDeliveryRepository(mysqlDB)
	deliverySvc := services.NewDeliveryService(deliveryRepo)

	// 5. Outbox Relay
	relay := outbox.NewRelay(deliveryRepo, publisher, 500*time.Millisecond, 50)
	go relay.Start(context.Background())

	// 6. Messaging Consumer
	idemHandler := commonMessaging.NewIdempotentHandler(deliveryRepo)
	consumer, err := messaging.NewRabbitMQConsumer(mqURL, deliverySvc, idemHandler)
	if err != nil {
		slog.Error("failed to connect to rabbitmq consumer", "error", err)
		os.Exit(1)
	}
	defer consumer.Close()

	if err := consumer.Start(context.Background()); err != nil {
		slog.Error("failed to start rabbitmq consumer", "error", err)
		os.Exit(1)
	}

	// 7. HTTP Server Setup
	e := api.NewEcho()
	courierHandler := api.NewCourierHandler(deliverySvc)
	courierHandler.RegisterRoutes(e)

	// 8. Start Server with Graceful Shutdown
	go func() {
		slog.Info("Starting Delivery HTTP server", "port", port)
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			slog.Error("shutting down the server", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		slog.Error("error shutting down http", "error", err)
	}

	slog.Info("Delivery service gracefully stopped")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
