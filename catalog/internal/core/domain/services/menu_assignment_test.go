package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/menu"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	common "github.com/vterry/food-ordering/common/pkg"
)

// ============================================================
// Helpers
// ============================================================

func validAddress() valueobjects.Address {
	addr, _ := valueobjects.NewAddress("Rua das Flores", "123", "", "Centro", "São Paulo", "SP", "01000-000")
	return addr
}

func createActiveMenu(restaurantID valueobjects.RestaurantID) *menu.Menu {
	m, _ := menu.NewMenu("MyMenu", restaurantID)
	cat, _ := menu.NewCategory("Drinks")
	item, _ := menu.NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
	_ = cat.AddItem(*item)
	_ = m.AddCategory(*cat)
	_ = m.Activate()
	return m
}

// ============================================================
// MenuAssignmentService
// ============================================================

func TestMenuAssignmentService_AssignMenuToRestaurant(t *testing.T) {
	svc := NewMenuAssignmentService()

	t.Run("assign active menu to its own restaurant succeeds", func(t *testing.T) {
		r, _ := restaurant.NewRestaurant("MyRestaurant", validAddress())
		m := createActiveMenu(r.RestaurantID)

		err := svc.AssignMenuToRestaurant(r, m)

		require.NoError(t, err)
		assert.Equal(t, m.MenuID, r.ActiveMenuID())
	})

	t.Run("cannot assign menu owned by another restaurant", func(t *testing.T) {
		r, _ := restaurant.NewRestaurant("MyRestaurant", validAddress())
		otherRestaurantID := valueobjects.NewRestaurantID()
		m := createActiveMenu(otherRestaurantID)

		err := svc.AssignMenuToRestaurant(r, m)

		assert.ErrorIs(t, err, ErrMenuNotOwnedByRestaurant)
	})

	t.Run("cannot assign draft menu", func(t *testing.T) {
		r, _ := restaurant.NewRestaurant("MyRestaurant", validAddress())
		m, _ := menu.NewMenu("DraftMenu", r.RestaurantID)

		err := svc.AssignMenuToRestaurant(r, m)

		assert.ErrorIs(t, err, ErrMenuNotActive)
	})

	t.Run("cannot assign archived menu", func(t *testing.T) {
		r, _ := restaurant.NewRestaurant("MyRestaurant", validAddress())
		m := createActiveMenu(r.RestaurantID)
		_ = m.Archive()

		err := svc.AssignMenuToRestaurant(r, m)

		assert.ErrorIs(t, err, ErrMenuNotActive)
	})
}
