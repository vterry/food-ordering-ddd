package ports

import (
	"context"

	"github.com/vterry/food-project/ordering/internal/core/domain/order"
	"github.com/vterry/food-project/ordering/internal/core/domain/saga"
)

type OrderRepository interface {
	Save(ctx context.Context, o *order.Order) error
	FindByID(ctx context.Context, id string) (*order.Order, error)
	UpdateWithVersion(ctx context.Context, o *order.Order, expectedVersion int) error
}

type SagaRepository interface {
	Save(ctx context.Context, s *saga.SagaState) error
	FindByOrderID(ctx context.Context, orderID string) (*saga.SagaState, error)
}

type CommandPublisher interface {
	PublishCommand(ctx context.Context, exchange, routingKey string, payload interface{}) error
}

type CustomerDTO struct {
	ID    string
	Name  string
	Email string
}

type CustomerClient interface {
	GetCustomerByID(ctx context.Context, customerID string) (*CustomerDTO, error)
}
