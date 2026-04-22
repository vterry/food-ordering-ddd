package ports

import (
	"context"

	"github.com/vterry/food-project/payment/internal/core/domain/payment"
)

type AuthorizeCommand struct {
	OrderID string
	Amount  int64
	Token   string
}

type CaptureCommand struct {
	PaymentID string
}

type RefundCommand struct {
	PaymentID string
}

type ReleaseCommand struct {
	PaymentID string
}

// PaymentUseCase defines the input port for the Payment service.
type PaymentUseCase interface {
	Authorize(ctx context.Context, cmd AuthorizeCommand) (*payment.Payment, error)
	Capture(ctx context.Context, cmd CaptureCommand) error
	Refund(ctx context.Context, cmd RefundCommand) error
	Release(ctx context.Context, cmd ReleaseCommand) error
	GetPayment(ctx context.Context, id string) (*payment.Payment, error)
	GetPaymentByOrderID(ctx context.Context, orderID string) (*payment.Payment, error)
}
