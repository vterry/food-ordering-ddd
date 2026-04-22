package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/vterry/food-project/payment/internal/core/domain/payment"
	"github.com/vterry/food-project/payment/internal/core/ports"
)

type paymentService struct {
	repo    ports.PaymentRepository
	gateway ports.PaymentGateway
}

func NewPaymentService(repo ports.PaymentRepository, gateway ports.PaymentGateway) *paymentService {
	return &paymentService{
		repo:    repo,
		gateway: gateway,
	}
}

func (s *paymentService) Authorize(ctx context.Context, cmd ports.AuthorizeCommand) (*payment.Payment, error) {
	// 1. Create new payment aggregate
	paymentID := uuid.New().String()
	p, err := payment.NewPayment(paymentID, cmd.OrderID, cmd.Amount)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment aggregate: %w", err)
	}

	token, err := payment.NewCardToken(cmd.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid card token: %w", err)
	}

	// 2. Try to authorize with external gateway
	err = s.gateway.Authorize(ctx, p, token)
	if err != nil {
		slog.Error("Payment authorization failed at gateway", "error", err, "order_id", cmd.OrderID)
		_ = p.FailAuthorization(err.Error())
	} else {
		_ = p.Authorize()
	}

	// 3. Save state (and outbox events)
	if err := s.repo.Save(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	return p, nil
}

func (s *paymentService) Capture(ctx context.Context, cmd ports.CaptureCommand) error {
	p, err := s.repo.FindByID(ctx, cmd.PaymentID)
	if err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("payment not found: %s", cmd.PaymentID)
	}

	// Try to capture funds
	err = s.gateway.Capture(ctx, p)
	if err != nil {
		slog.Error("Payment capture failed at gateway", "error", err, "payment_id", p.ID())
		// In a real system, we might retry or mark it for manual intervention
		return fmt.Errorf("gateway capture failed: %w", err)
	}

	if err := p.Capture(); err != nil {
		return err
	}

	return s.repo.Save(ctx, p)
}

func (s *paymentService) Refund(ctx context.Context, cmd ports.RefundCommand) error {
	p, err := s.repo.FindByID(ctx, cmd.PaymentID)
	if err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("payment not found: %s", cmd.PaymentID)
	}

	err = s.gateway.Refund(ctx, p)
	if err != nil {
		return fmt.Errorf("gateway refund failed: %w", err)
	}

	if err := p.Refund(); err != nil {
		return err
	}

	return s.repo.Save(ctx, p)
}

func (s *paymentService) Release(ctx context.Context, cmd ports.ReleaseCommand) error {
	p, err := s.repo.FindByID(ctx, cmd.PaymentID)
	if err != nil {
		return err
	}
	if p == nil {
		return fmt.Errorf("payment not found: %s", cmd.PaymentID)
	}

	err = s.gateway.Release(ctx, p)
	if err != nil {
		return fmt.Errorf("gateway release failed: %w", err)
	}

	if err := p.Release(); err != nil {
		return err
	}

	return s.repo.Save(ctx, p)
}

func (s *paymentService) GetPayment(ctx context.Context, id string) (*payment.Payment, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *paymentService) GetPaymentByOrderID(ctx context.Context, orderID string) (*payment.Payment, error) {
	return s.repo.FindByOrderID(ctx, orderID)
}
