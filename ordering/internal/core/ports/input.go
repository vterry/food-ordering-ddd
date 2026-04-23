package ports

import (
	"context"
	"github.com/vterry/food-project/ordering/internal/core/domain/order"
)

type CreateOrderCommand struct {
	CustomerID    string
	RestaurantID  string
	Items         []OrderItemDTO
	CardToken     string
	CorrelationID string
}

type OrderItemDTO struct {
	ProductID string
	Name      string
	Quantity  int
	Price     float64
}

type OrderUseCase interface {
	CreateOrder(ctx context.Context, cmd CreateOrderCommand) (string, error)
	CancelOrder(ctx context.Context, orderID string) error
	GetOrder(ctx context.Context, orderID string) (*order.Order, error)
}
