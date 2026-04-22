package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewEchoServer(handler *PaymentHandler) *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	v1 := e.Group("/api/v1")
	v1.GET("/payments/:id", handler.GetPayment)

	e.GET("/health/live", handler.HealthLive)
	e.GET("/health/ready", handler.HealthReady)

	return e
}
