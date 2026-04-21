package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/vterry/food-project/customer/internal/adapters/api"
	apigen "github.com/vterry/food-project/customer/internal/adapters/api/generated"
	"github.com/vterry/food-project/customer/internal/adapters/api/middleware"
	customerDB "github.com/vterry/food-project/customer/internal/adapters/db"
	"github.com/vterry/food-project/customer/internal/adapters/db/repository"
	"github.com/vterry/food-project/customer/internal/adapters/messaging"
	"github.com/vterry/food-project/customer/internal/core/services"
)

const (
	defaultPort = "8080"
	defaultDSN  = "root:root@tcp(localhost:3306)/food_project?parseTime=true"
	defaultMQ   = "amqp://guest:guest@localhost:5672/"
)

func main() {
	// 1. Config (Basic env load)
	port := getEnv("PORT", defaultPort)
	dsn := getEnv("MYSQL_DSN", defaultDSN)
	mqURL := getEnv("RABBITMQ_URL", defaultMQ)

	// 2. Infra Initialization
	mysqlDB, err := customerDB.NewMySQLConnection(dsn)
	if err != nil {
		log.Fatalf("failed to connect to mysql: %v", err)
	}
	defer mysqlDB.Close()

	publisher, err := messaging.NewRabbitMQPublisher(mqURL, "customer-service")
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer publisher.Close()

	// 3. Dependency Injection
	custRepo := repository.NewSQLCustomerRepository(mysqlDB)
	cartRepo := repository.NewSQLCartRepository(mysqlDB)

	custSvc := services.NewCustomerService(custRepo, publisher)
	cartSvc := services.NewCartService(cartRepo, publisher)

	handler := api.NewCustomerHandler(custSvc, cartSvc)

	// 4. Server Setup
	e := echo.New()
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(middleware.CorrelationIDMiddleware)

	// Register Handlers
	strictHandler := apigen.NewStrictHandler(handler, nil)
	apigen.RegisterHandlersWithBaseURL(e, strictHandler, "/api/v1")

	// 5. Start Server with Graceful Shutdown
	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	
	fmt.Println("Server gracefully stopped")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
