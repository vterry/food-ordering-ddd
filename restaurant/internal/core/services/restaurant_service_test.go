package services

import (
	"context"
	"testing"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/restaurant/internal/core/domain/restaurant"
	"github.com/vterry/food-project/restaurant/internal/core/ports"
)

func TestRestaurantService_CreateRestaurant(t *testing.T) {
	repo := NewMockRestaurantRepo()
	svc := NewRestaurantService(repo, nil)

	cmd := ports.CreateRestaurantCommand{
		Name:    "Gopher Grill",
		Address: restaurant.Address{Street: "123 Go Lane", City: "Dev City", ZipCode: "00000"},
	}

	id, err := svc.CreateRestaurant(context.Background(), cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rest, _ := repo.FindByID(context.Background(), id)
	if rest == nil || rest.Name() != "Gopher Grill" {
		t.Error("restaurant was not saved correctly")
	}
}

func TestRestaurantService_ActivateMenuConformingInvariant(t *testing.T) {
	mRepo := NewMockMenuRepo()
	svc := NewRestaurantService(nil, mRepo)
	restaurantID := vo.NewID("r1")

	// Create 2 menus
	m1ID, _ := svc.CreateMenu(context.Background(), restaurantID, "Menu 1")
	m2ID, _ := svc.CreateMenu(context.Background(), restaurantID, "Menu 2")

	// Activate Menu 1
	_ = svc.ActivateMenu(context.Background(), m1ID)
	
	m1, _ := mRepo.FindByID(context.Background(), m1ID)
	if !m1.IsActive() {
		t.Error("Menu 1 should be active")
	}

	// Activate Menu 2 -> Should deactivate Menu 1
	_ = svc.ActivateMenu(context.Background(), m2ID)

	m1, _ = mRepo.FindByID(context.Background(), m1ID)
	m2, _ := mRepo.FindByID(context.Background(), m2ID)

	if m1.IsActive() {
		t.Error("Menu 1 should have been deactivated")
	}
	if !m2.IsActive() {
		t.Error("Menu 2 should be active")
	}
}
