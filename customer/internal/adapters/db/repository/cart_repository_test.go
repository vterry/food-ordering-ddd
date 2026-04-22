package repository

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/customer/internal/core/domain/cart"
	"testing"
)

func TestSQLCartRepository_Integration(t *testing.T) {
	skipIfShort(t)
	db, teardown := setupMySQLContainer(t)
	defer teardown()

	repo := NewSQLCartRepository(db)
	ctx := context.Background()

	t.Run("Save and FindByCustomerID", func(t *testing.T) {
		customerID := vo.NewID("cust-cart-1")
		restID := vo.NewID("rest-1")
		c := cart.NewCart(vo.NewID("cart-1"), customerID)

		money, _ := vo.NewMoney(15.0, "BRL")
		item := cart.NewCartItem(vo.NewID("prod-1"), "Burger", money, 2, "no onions")
		_ = c.AddItem(restID, item)

		err := repo.Save(ctx, c)
		if err != nil {
			t.Fatalf("failed to save cart: %v", err)
		}

		got, err := repo.FindByCustomerID(ctx, customerID)
		if err != nil {
			t.Fatalf("failed to find cart: %v", err)
		}
		if got == nil {
			t.Fatal("expected cart, got nil")
		}
		if !got.RestaurantID().Equals(restID) {
			t.Errorf("expected restaurant %v, got %v", restID, got.RestaurantID())
		}
		if len(got.Items()) != 1 {
			t.Errorf("expected 1 item, got %d", len(got.Items()))
		}
		if got.Items()[0].Name() != "Burger" {
			t.Errorf("expected item Burger, got %s", got.Items()[0].Name())
		}
		if got.Items()[0].Observation() != "no onions" {
			t.Errorf("expected observation 'no onions', got '%s'", got.Items()[0].Observation())
		}
	})

	t.Run("Update Cart Items", func(t *testing.T) {
		customerID := vo.NewID("cust-cart-1")
		c, _ := repo.FindByCustomerID(ctx, customerID)
		
		c.UpdateItemQuantity(c.Items()[0].ProductID(), 5)
		
		err := repo.Save(ctx, c)
		if err != nil {
			t.Fatalf("failed to update cart: %v", err)
		}

		got, _ := repo.FindByCustomerID(ctx, customerID)
		if got.Items()[0].Quantity() != 5 {
			t.Errorf("expected quantity 5, got %d", got.Items()[0].Quantity())
		}
	})

	t.Run("Delete Cart", func(t *testing.T) {
		customerID := vo.NewID("cust-cart-1")
		err := repo.Delete(ctx, customerID)
		if err != nil {
			t.Fatalf("failed to delete cart: %v", err)
		}

		got, _ := repo.FindByCustomerID(ctx, customerID)
		if got != nil {
			t.Error("expected cart to be deleted")
		}
	})
}
