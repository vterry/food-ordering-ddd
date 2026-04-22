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

	"github.com/vterry/food-project/payment/internal/adapters/api"
	"github.com/vterry/food-project/payment/internal/adapters/db"
	"github.com/vterry/food-project/payment/internal/adapters/db/repository"
	"github.com/vterry/food-project/payment/internal/adapters/external"
	"github.com/vterry/food-project/payment/internal/adapters/messaging"
	"github.com/vterry/food-project/payment/internal/core/services"
)

func main() {
	// Configuration (In a real app, use env vars)
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbDSN = "root:root@tcp(localhost:3306)/payment?parseTime=true"
	}
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	// 1. Inbound/Outbound Adapters
	database, err := db.NewDB(dbDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	repo := repository.NewPaymentRepository(database)
	gateway := external.NewMockGateway()

	// 2. Application Service
	paymentService := services.NewPaymentService(repo, gateway)

	// 3. Inbound Adapters (API & Messaging)
	paymentHandler := api.NewPaymentHandler(paymentService)
	server := api.NewEchoServer(paymentHandler)

	consumer, err := messaging.NewRabbitMQConsumer(rabbitURL, paymentService)
	if err != nil {
		slog.Warn("Failed to initialize RabbitMQ consumer", "error", err)
	} else {
		defer consumer.Close()
		if err := consumer.Start(context.Background()); err != nil {
			log.Fatalf("Failed to start RabbitMQ consumer: %v", err)
		}
	}

	// 4. Start Server
	go func() {
		if err := server.Start(":8082"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 5. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	slog.Info("Server stopped")
}
