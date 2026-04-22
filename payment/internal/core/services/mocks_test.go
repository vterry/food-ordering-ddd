package services

import (
	"context"
	"github.com/vterry/food-project/payment/internal/core/domain/payment"
)

type MockPaymentRepo struct {
	payments   map[string]*payment.Payment
	SaveCalled int
	SaveErr    error
	FindErr    error
}

func NewMockPaymentRepo() *MockPaymentRepo {
	return &MockPaymentRepo{payments: make(map[string]*payment.Payment)}
}

func (m *MockPaymentRepo) Save(ctx context.Context, p *payment.Payment) error {
	m.SaveCalled++
	if m.SaveErr != nil {
		return m.SaveErr
	}
	m.payments[p.ID()] = p
	return nil
}

func (m *MockPaymentRepo) FindByID(ctx context.Context, id string) (*payment.Payment, error) {
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	return m.payments[id], nil
}

func (m *MockPaymentRepo) FindByOrderID(ctx context.Context, orderID string) (*payment.Payment, error) {
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	for _, p := range m.payments {
		if p.OrderID() == orderID {
			return p, nil
		}
	}
	return nil, nil
}

type MockGateway struct {
	AuthorizeErr error
	CaptureErr   error
	RefundErr    error
	ReleaseErr   error
}

func (m *MockGateway) Authorize(ctx context.Context, p *payment.Payment, token payment.CardToken) error {
	return m.AuthorizeErr
}

func (m *MockGateway) Capture(ctx context.Context, p *payment.Payment) error {
	return m.CaptureErr
}

func (m *MockGateway) Refund(ctx context.Context, p *payment.Payment) error {
	return m.RefundErr
}

func (m *MockGateway) Release(ctx context.Context, p *payment.Payment) error {
	return m.ReleaseErr
}
