package payment

import (
	"errors"
	"testing"
)

func TestNewPayment(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		orderID string
		amount  int64
		wantErr bool
	}{
		{"Valid payment", "p1", "o1", 1000, false},
		{"Zero amount", "p2", "o2", 0, true},
		{"Negative amount", "p3", "o3", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPayment(tt.id, tt.orderID, tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPayment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if p.ID() != tt.id {
					t.Errorf("got id %v, want %v", p.ID(), tt.id)
				}
				if p.Status() != StatusCreated {
					t.Errorf("got status %v, want %v", p.Status(), StatusCreated)
				}
			}
		})
	}
}

func TestPayment_Transitions(t *testing.T) {
	t.Run("Full happy path: Authorize -> Capture", func(t *testing.T) {
		p, _ := NewPayment("p1", "o1", 1000)
		if err := p.Authorize(); err != nil {
			t.Fatalf("Authorize failed: %v", err)
		}
		if p.Status() != StatusAuthorized {
			t.Errorf("expected AUTHORIZED, got %s", p.Status())
		}

		if err := p.Capture(); err != nil {
			t.Fatalf("Capture failed: %v", err)
		}
		if p.Status() != StatusCaptured {
			t.Errorf("expected CAPTURED, got %s", p.Status())
		}

		events := p.PullEvents()
		if len(events) != 2 {
			t.Errorf("expected 2 events, got %d", len(events))
		}
	})

	t.Run("Compensation: Authorize -> Release", func(t *testing.T) {
		p, _ := NewPayment("p1", "o1", 1000)
		p.Authorize()
		if err := p.Release(); err != nil {
			t.Fatalf("Release failed: %v", err)
		}
		if p.Status() != StatusReleased {
			t.Errorf("expected RELEASED, got %s", p.Status())
		}
	})

	t.Run("Compensation: Capture -> Refund", func(t *testing.T) {
		p, _ := NewPayment("p1", "o1", 1000)
		p.Authorize()
		p.Capture()
		if err := p.Refund(); err != nil {
			t.Fatalf("Refund failed: %v", err)
		}
		if p.Status() != StatusRefunded {
			t.Errorf("expected REFUNDED, got %s", p.Status())
		}
	})

	t.Run("Failed Authorization", func(t *testing.T) {
		p, _ := NewPayment("p1", "o1", 1000)
		if err := p.FailAuthorization("insufficient funds"); err != nil {
			t.Fatalf("FailAuthorization failed: %v", err)
		}
		if p.Status() != StatusAuthorizationFailed {
			t.Errorf("expected AUTHORIZATION_FAILED, got %s", p.Status())
		}
	})

	t.Run("Idempotency", func(t *testing.T) {
		p, _ := NewPayment("p1", "o1", 1000)
		p.Authorize()
		err := p.Authorize()
		if !errors.Is(err, ErrPaymentAlreadyProcessed) {
			t.Errorf("expected ErrPaymentAlreadyProcessed, got %v", err)
		}
	})

	t.Run("Invalid transition", func(t *testing.T) {
		p, _ := NewPayment("p1", "o1", 1000)
		// Try to capture before authorizing
		err := p.Capture()
		if !errors.Is(err, ErrInvalidStateTransition) {
			t.Errorf("expected ErrInvalidStateTransition, got %v", err)
		}
	})
}
