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

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/vterry/food-ordering/catalog/internal/infra/config"
	"github.com/vterry/food-ordering/catalog/internal/infra/db/mysql"
	"github.com/vterry/food-ordering/catalog/internal/infra/server"
)

func main() {
	cfg := config.NewConfig()

	database, err := mysql.NewDataBase(cfg.Db)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Printf("Database connection established")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	rabbitCon, err := amqp.Dial(cfg.OutboxCfg.Address)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	rabbitChan, err := rabbitCon.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
	}

	logger.Info("RabbitMQ connection established")

	server := server.NewCatalogServer(cfg, database, rabbitCon, rabbitChan, logger)

	go func() {
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server failed: %v", err)
		}
	}()

	select {
	case <-server.NotifyReady:
		log.Println("Catalog server started")
	case <-time.After(5 * time.Second):
		log.Println("Warning: Server startup timed out or is taking longer than expected")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")

}
