package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vterry/food-project/payment/internal/core/ports"
)

type PaymentHandler struct {
	paymentService ports.PaymentUseCase
}

func NewPaymentHandler(paymentService ports.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
	}
}

func (h *PaymentHandler) GetPayment(c echo.Context) error {
	id := c.Param("id")
	p, err := h.paymentService.GetPayment(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if p == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "payment not found"})
	}

	return c.JSON(http.StatusOK, p)
}

func (h *PaymentHandler) HealthLive(c echo.Context) error {
	return c.String(http.StatusOK, "live")
}

func (h *PaymentHandler) HealthReady(c echo.Context) error {
	// In a real system, check DB and RabbitMQ connections
	return c.String(http.StatusOK, "ready")
}
