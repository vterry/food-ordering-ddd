package repository

import (
	"context"
	"testing"

	"github.com/vterry/food-project/payment/internal/core/domain/payment"
)

func TestPaymentRepository_SaveAndFind(t *testing.T) {
	skipIfShort(t)
	db, teardown := setupMySQLContainer(t)
	defer teardown()

	repo := NewPaymentRepository(db)
	ctx := context.Background()

	t.Run("Create and FindByID", func(t *testing.T) {
		p, _ := payment.NewPayment("p1", "o1", 5000)
		_ = p.Authorize()

		err := repo.Save(ctx, p)
		if err != nil {
			t.Fatalf("failed to save payment: %v", err)
		}

		got, err := repo.FindByID(ctx, "p1")
		if err != nil {
			t.Fatalf("failed to find payment: %v", err)
		}

		if got == nil {
			t.Fatal("expected payment, got nil")
		}
		if got.ID() != "p1" {
			t.Errorf("expected id p1, got %s", got.ID())
		}
		if got.Status() != payment.StatusAuthorized {
			t.Errorf("expected status AUTHORIZED, got %s", got.Status())
		}
	})

	t.Run("Update and FindByOrderID", func(t *testing.T) {
		p, _ := repo.FindByID(ctx, "p1")
		_ = p.Capture()

		err := repo.Save(ctx, p)
		if err != nil {
			t.Fatalf("failed to update payment: %v", err)
		}

		got, err := repo.FindByOrderID(ctx, "o1")
		if err != nil {
			t.Fatalf("failed to find payment by order: %v", err)
		}

		if got.Status() != payment.StatusCaptured {
			t.Errorf("expected status CAPTURED, got %s", got.Status())
		}
	})
}
