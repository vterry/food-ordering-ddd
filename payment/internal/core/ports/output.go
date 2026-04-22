package ports

import (
	"context"

	"github.com/vterry/food-project/payment/internal/core/domain/payment"
)

// PaymentGateway is the outbound port for interacting with external payment providers.
type PaymentGateway interface {
	Authorize(ctx context.Context, p *payment.Payment, token payment.CardToken) error
	Capture(ctx context.Context, p *payment.Payment) error
	Refund(ctx context.Context, p *payment.Payment) error
	Release(ctx context.Context, p *payment.Payment) error
}

// PaymentRepository is the outbound port for interacting with the database.
type PaymentRepository interface {
	Save(ctx context.Context, p *payment.Payment) error
	FindByID(ctx context.Context, id string) (*payment.Payment, error)
	FindByOrderID(ctx context.Context, orderID string) (*payment.Payment, error)
}
