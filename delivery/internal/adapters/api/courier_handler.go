package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/delivery/internal/core/domain/delivery"
	"github.com/vterry/food-project/delivery/internal/core/ports"
)

type CourierHandler struct {
	useCase ports.DeliveryUseCase
}

func NewCourierHandler(useCase ports.DeliveryUseCase) *CourierHandler {
	return &CourierHandler{useCase: useCase}
}

func (h *CourierHandler) RegisterRoutes(e *echo.Echo) {
	g := e.Group("/deliveries/:id")
	g.POST("/pickup", h.PickUp)
	g.POST("/complete", h.Complete)
	g.POST("/refuse", h.Refuse)
}

type PickUpRequest struct {
	CourierID   string `json:"courier_id"`
	CourierName string `json:"courier_name"`
}

func (h *CourierHandler) PickUp(c echo.Context) error {
	id := vo.NewID(c.Param("id"))
	var req PickUpRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	courier, err := delivery.NewCourierInfo(req.CourierID, req.CourierName)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.useCase.PickUpDelivery(c.Request().Context(), id, courier); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *CourierHandler) Complete(c echo.Context) error {
	id := vo.NewID(c.Param("id"))
	if err := h.useCase.CompleteDelivery(c.Request().Context(), id); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}

type RefuseRequest struct {
	Reason string `json:"reason"`
}

func (h *CourierHandler) Refuse(c echo.Context) error {
	id := vo.NewID(c.Param("id"))
	var req RefuseRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.useCase.RefuseDelivery(c.Request().Context(), id, req.Reason); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
