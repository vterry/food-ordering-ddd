package api

import (
	"context"

	"github.com/google/uuid"
	"github.com/vterry/food-project/ordering/internal/adapters/api/generated"
	"github.com/vterry/food-project/ordering/internal/core/ports"
)

type OrderHandler struct {
	service ports.OrderUseCase
}

func NewOrderHandler(service ports.OrderUseCase) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, request generated.CreateOrderRequestObject) (generated.CreateOrderResponseObject, error) {
	items := make([]ports.OrderItemDTO, 0, len(request.Body.Items))
	for _, it := range request.Body.Items {
		items = append(items, ports.OrderItemDTO{
			ProductID: it.ProductId,
			Name:      it.Name,
			Quantity:  it.Quantity,
			Price:     float64(it.Price),
		})
	}

	cmd := ports.CreateOrderCommand{
		CustomerID:    request.Body.CustomerId,
		RestaurantID:  request.Body.RestaurantId,
		Items:         items,
		CardToken:     request.Body.CardToken,
		CorrelationID: uuid.New().String(),
	}

	id, err := h.service.CreateOrder(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return generated.CreateOrder202JSONResponse{
		Id:     &id,
		Status: stringPtr("CREATED"),
	}, nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, request generated.GetOrderRequestObject) (generated.GetOrderResponseObject, error) {
	o, err := h.service.GetOrder(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return generated.GetOrder404Response{}, nil
	}

	amount := float32(o.TotalAmount().Amount())
	createdAt := o.CreatedAt()

	return generated.GetOrder200JSONResponse{
		Id:           stringPtr(o.ID().String()),
		CustomerId:   stringPtr(o.CustomerID().String()),
		RestaurantId: stringPtr(o.RestaurantID().String()),
		Status:       stringPtr(o.Status().String()),
		TotalAmount:  &amount,
		CreatedAt:    &createdAt,
	}, nil
}

func (h *OrderHandler) CancelOrder(ctx context.Context, request generated.CancelOrderRequestObject) (generated.CancelOrderResponseObject, error) {
	err := h.service.CancelOrder(ctx, request.Id)
	if err != nil {
		return generated.CancelOrder400Response{}, nil
	}
	return generated.CancelOrder204Response{}, nil
}

func stringPtr(s string) *string {
	return &s
}
