package repository

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/restaurant/internal/core/domain/restaurant"
	"testing"
)

func TestSQLRestaurantRepository_Integration(t *testing.T) {
	skipIfShort(t)
	db, teardown := setupMySQLContainer(t)
	defer teardown()

	repo := NewSQLRestaurantRepository(db)
	ctx := context.Background()

	t.Run("Save and FindByID", func(t *testing.T) {
		id := vo.NewID("rest-1")
		addr := restaurant.Address{Street: "123 Go Blvd", City: "Dev", ZipCode: "12345"}
		hours := []restaurant.OperatingPeriod{
			{DayOfWeek: 1, Open: "08:00", Close: "22:00"},
		}
		rest := restaurant.NewRestaurant(id, "The Gopher", addr, hours)

		err := repo.Save(ctx, rest)
		if err != nil {
			t.Fatalf("failed to save restaurant: %v", err)
		}

		got, err := repo.FindByID(ctx, id)
		if err != nil {
			t.Fatalf("failed to find restaurant: %v", err)
		}
		if got == nil {
			t.Fatal("expected restaurant, got nil")
		}
		if got.Name() != "The Gopher" {
			t.Errorf("expected name The Gopher, got %s", got.Name())
		}
		if len(got.OperatingHours()) != 1 {
			t.Errorf("expected 1 operating period, got %d", len(got.OperatingHours()))
		}
	})
}
