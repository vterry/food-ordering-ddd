package repository

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/restaurant/internal/core/domain/menu"
	"github.com/vterry/food-project/restaurant/internal/core/domain/restaurant"
	"testing"
)

func TestSQLMenuRepository_Integration(t *testing.T) {
	skipIfShort(t)
	db, teardown := setupMySQLContainer(t)
	defer teardown()

	repo := NewSQLMenuRepository(db)
	ctx := context.Background()

	// Need a restaurant first due to FK
	restRepo := NewSQLRestaurantRepository(db)
	restID := vo.NewID("r1")
	// Use domain factory
	addr := restaurant.Address{Street: "S", City: "C", ZipCode: "Z"}
	rest := restaurant.NewRestaurant(restID, "Test Restaurant", addr, nil)
	_ = restRepo.Save(ctx, rest)

	t.Run("Save and FindActive", func(t *testing.T) {
		id := vo.NewID("menu-1")
		m := menu.NewMenu(id, restID, "Summer Menu")
		m.Activate()

		price, _ := vo.NewMoney(10.0, "USD")
		item := menu.NewMenuItem(vo.NewID("item-1"), "Burger", "Tasty", price, "Mains")
		m.AddItem(item)

		err := repo.Save(ctx, m)
		if err != nil {
			t.Fatalf("failed to save menu: %v", err)
		}

		got, err := repo.FindActiveByRestaurantID(ctx, restID)
		if err != nil {
			t.Fatalf("failed to find active menu: %v", err)
		}
		if got == nil || got.ID().String() != "menu-1" {
			t.Fatal("expected active menu-1")
		}
		if len(got.Items()) != 1 {
			t.Errorf("expected 1 item, got %d", len(got.Items()))
		}
	})
}
