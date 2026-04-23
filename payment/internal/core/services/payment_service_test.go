package services

import (
	"context"
	"errors"
	"testing"

	"github.com/vterry/food-project/payment/internal/core/domain/payment"
	"github.com/vterry/food-project/payment/internal/core/ports"
)

func TestPaymentService_Authorize(t *testing.T) {
	t.Run("successful authorization", func(t *testing.T) {
		repo := NewMockPaymentRepo()
		gateway := &MockGateway{}
		svc := NewPaymentService(repo, gateway, &MockPublisher{})

		cmd := ports.AuthorizeCommand{
			OrderID: "ord-1",
			Amount:  1000,
			Token:   "valid-token",
		}

		p, err := svc.Authorize(context.Background(), cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if p.Status() != payment.StatusAuthorized {
			t.Errorf("expected AUTHORIZED, got %s", p.Status())
		}
		if repo.SaveCalled != 1 {
			t.Errorf("expected repo.Save to be called once, got %d", repo.SaveCalled)
		}
	})

	t.Run("gateway failure", func(t *testing.T) {
		repo := NewMockPaymentRepo()
		gateway := &MockGateway{AuthorizeErr: errors.New("gateway rejected")}
		svc := NewPaymentService(repo, gateway, &MockPublisher{})

		cmd := ports.AuthorizeCommand{
			OrderID: "ord-2",
			Amount:  2000,
			Token:   "invalid-token",
		}

		p, err := svc.Authorize(context.Background(), cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if p.Status() != payment.StatusAuthorizationFailed {
			t.Errorf("expected AUTHORIZATION_FAILED, got %s", p.Status())
		}
	})
}

func TestPaymentService_Capture(t *testing.T) {
	t.Run("successful capture", func(t *testing.T) {
		repo := NewMockPaymentRepo()
		p, _ := payment.NewPayment("p1", "o1", 1000)
		p.Authorize()
		_ = repo.Save(context.Background(), p)
		repo.SaveCalled = 0 // Reset counter

		gateway := &MockGateway{}
		svc := NewPaymentService(repo, gateway, &MockPublisher{})

		err := svc.Capture(context.Background(), ports.CaptureCommand{PaymentID: "p1"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if p.Status() != payment.StatusCaptured {
			t.Errorf("expected CAPTURED, got %s", p.Status())
		}
	})

	t.Run("payment not found", func(t *testing.T) {
		repo := NewMockPaymentRepo()
		svc := NewPaymentService(repo, &MockGateway{}, &MockPublisher{})

		err := svc.Capture(context.Background(), ports.CaptureCommand{PaymentID: "unknown"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
