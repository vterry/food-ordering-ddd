package payment

import (
	"errors"
	"fmt"
	"time"
	"github.com/vterry/food-project/common/pkg/domain/base"
)

var (
	ErrInvalidStateTransition  = errors.New("invalid state transition")
	ErrPaymentAlreadyProcessed = errors.New("payment already processed (idempotent)")
)

// Payment is the aggregate root for the Payment bounded context.
type Payment struct {
	id        string
	orderID   string
	amount    int64 // in cents
	status    Status
	events    []base.DomainEvent
	updatedAt time.Time
}

// NewPayment creates a new payment aggregate in the CREATED state.
func NewPayment(id, orderID string, amount int64) (*Payment, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	return &Payment{
		id:        id,
		orderID:   orderID,
		amount:    amount,
		status:    StatusCreated,
		updatedAt: time.Now(),
	}, nil
}

// MapFromPersistence reconstructs the aggregate from database values.
func MapFromPersistence(id, orderID string, amount int64, status Status, updatedAt time.Time) *Payment {
	return &Payment{
		id:        id,
		orderID:   orderID,
		amount:    amount,
		status:    status,
		updatedAt: updatedAt,
	}
}

// Authorize transition the payment to AUTHORIZED.
func (p *Payment) Authorize() error {
	if p.status == StatusAuthorized {
		return ErrPaymentAlreadyProcessed
	}
	if p.status != StatusCreated {
		return fmt.Errorf("%w: cannot authorize payment from status %s", ErrInvalidStateTransition, p.status)
	}
	p.status = StatusAuthorized
	p.updatedAt = time.Now()
	p.addEvent(PaymentAuthorized{
		PaymentID:      p.id,
		OrderID:        p.orderID,
		Amount:         p.amount,
		OccurredAtTime: p.updatedAt,
	})
	return nil
}

// FailAuthorization transition the payment to AUTHORIZATION_FAILED.
func (p *Payment) FailAuthorization(reason string) error {
	if p.status == StatusAuthorizationFailed {
		return ErrPaymentAlreadyProcessed
	}
	if p.status != StatusCreated {
		return fmt.Errorf("%w: cannot fail authorization from status %s", ErrInvalidStateTransition, p.status)
	}
	p.status = StatusAuthorizationFailed
	p.updatedAt = time.Now()
	p.addEvent(PaymentAuthorizationFailed{
		PaymentID:      p.id,
		OrderID:        p.orderID,
		Reason:         reason,
		OccurredAtTime: p.updatedAt,
	})
	return nil
}

// Capture transition the payment to CAPTURED.
func (p *Payment) Capture() error {
	if p.status == StatusCaptured {
		return ErrPaymentAlreadyProcessed
	}
	if p.status != StatusAuthorized {
		return fmt.Errorf("%w: cannot capture payment from status %s", ErrInvalidStateTransition, p.status)
	}
	p.status = StatusCaptured
	p.updatedAt = time.Now()
	p.addEvent(PaymentCaptured{
		PaymentID:      p.id,
		OrderID:        p.orderID,
		Amount:         p.amount,
		OccurredAtTime: p.updatedAt,
	})
	return nil
}

// Release transition the payment to RELEASED (voiding an authorization).
func (p *Payment) Release() error {
	if p.status == StatusReleased {
		return ErrPaymentAlreadyProcessed
	}
	if p.status != StatusAuthorized {
		return fmt.Errorf("%w: cannot release payment from status %s", ErrInvalidStateTransition, p.status)
	}
	p.status = StatusReleased
	p.updatedAt = time.Now()
	p.addEvent(PaymentReleased{
		PaymentID:      p.id,
		OrderID:        p.orderID,
		OccurredAtTime: p.updatedAt,
	})
	return nil
}

// Refund transition the payment to REFUNDED.
func (p *Payment) Refund() error {
	if p.status == StatusRefunded {
		return ErrPaymentAlreadyProcessed
	}
	if p.status != StatusCaptured {
		return fmt.Errorf("%w: cannot refund payment from status %s", ErrInvalidStateTransition, p.status)
	}
	p.status = StatusRefunded
	p.updatedAt = time.Now()
	p.addEvent(PaymentRefunded{
		PaymentID:      p.id,
		OrderID:        p.orderID,
		Amount:         p.amount,
		OccurredAtTime: p.updatedAt,
	})
	return nil
}

// Getters

func (p *Payment) ID() string      { return p.id }
func (p *Payment) OrderID() string { return p.orderID }
func (p *Payment) Amount() int64   { return p.amount }
func (p *Payment) Status() Status  { return p.status }

func (p *Payment) addEvent(e base.DomainEvent) {
	p.events = append(p.events, e)
}

// PullEvents returns all collected domain events and clears the internal slice.
func (p *Payment) PullEvents() []base.DomainEvent {
	events := p.events
	p.events = nil
	return events
}

// GetEvents returns all collected domain events without clearing them.
func (p *Payment) GetEvents() []base.DomainEvent {
	return p.events
}
