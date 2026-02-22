package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/menu"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	common "github.com/vterry/food-ordering/common/pkg"
)

func TestMenuRepository_Save(t *testing.T) {
	repo := NewMenuRepository(testDB, NewOutboxRepository(testDB))

	testMenuName := "Test Menu"
	testCatName := "Beverages"
	testItemName := "TestItemName"
	testItemPrice := common.NewMoneyFromCents(12345)

	tests := []struct {
		name       string
		setup      func(t *testing.T) *menu.Menu
		wantErr    bool
		assertFunc func(t *testing.T, id valueobjects.MenuID)
	}{
		{
			name: "save new menu with items",
			setup: func(t *testing.T) *menu.Menu {
				addr, err := valueobjects.NewAddress("Test Street", "123", "", "District", "City", "ST", "12345-2")
				require.NoError(t, err)

				r, err := restaurant.NewRestaurant("TestRest", addr)
				require.NoError(t, err)

				insertRestaurant(t, r)

				m, err := menu.NewMenu(testMenuName, r.RestaurantID)
				require.NoError(t, err)

				cat, _ := menu.NewCategory(testCatName)
				_ = m.AddCategory(*cat)

				item, _ := menu.NewItemMenu(testItemName, "Test description", testItemPrice)
				_ = m.AddItemToCategory(cat.CategoryID, *item)

				return m
			},
			wantErr: false,
			assertFunc: func(t *testing.T, id valueobjects.MenuID) {
				foundMenu, err := repo.FindById(context.Background(), id)
				require.NoError(t, err)
				require.NotNil(t, foundMenu)

				assert.Equal(t, id, foundMenu.MenuID)
				assert.Equal(t, testMenuName, foundMenu.Name())

				cats := foundMenu.Categories()
				require.Len(t, cats, 1)

				cat := cats[0]
				assert.Equal(t, testCatName, cat.Name())

				items := cat.Items()
				require.Len(t, items, 1)

				item := items[0]
				assert.Equal(t, testItemName, item.Name())

				assert.True(t, testItemPrice.Amount() == item.BasePrice().Amount())

				eventCount := CountEventsInOutbox(t, testDB, id.String())
				assert.Equal(t, 3, eventCount)

				lastPayload := GetLastEventPayload(t, testDB, id.String())
				assert.NotNil(t, lastPayload)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateTables(testDB)
			menuAgg := tt.setup(t)

			err := repo.Save(context.Background(), menuAgg)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tt.assertFunc != nil {
				tt.assertFunc(t, menuAgg.MenuID)
			}
		})
	}
}

func TestMenuRepository_FindByRestaurantId(t *testing.T) {
	repo := NewMenuRepository(testDB, NewOutboxRepository(testDB))

	t.Run("find menus by restaurant id", func(t *testing.T) {
		truncateTables(testDB)

		// Setup
		addr, err := valueobjects.NewAddress("Test Street", "123", "", "District", "City", "ST", "12345-2")
		require.NoError(t, err)

		r1, err := restaurant.NewRestaurant("Rest One", addr)
		require.NoError(t, err)

		r2, err := restaurant.NewRestaurant("Rest Two", addr)
		require.NoError(t, err)

		insertRestaurant(t, r1)
		insertRestaurant(t, r2)

		m1, err := menu.NewMenu("Menu 1 R1", r1.RestaurantID)
		require.NoError(t, err)
		m2, err := menu.NewMenu("Menu 2 R1", r1.RestaurantID)
		require.NoError(t, err)
		m3, err := menu.NewMenu("Menu 1 R2", r2.RestaurantID)
		require.NoError(t, err)

		err = repo.Save(context.Background(), m1)
		require.NoError(t, err)
		err = repo.Save(context.Background(), m2)
		require.NoError(t, err)
		err = repo.Save(context.Background(), m3)
		require.NoError(t, err)

		// Act
		menus, err := repo.FindByRestaurantId(context.Background(), r1.RestaurantID)

		// Assert
		require.NoError(t, err)
		assert.Len(t, menus, 2)
		for _, m := range menus {
			assert.Equal(t, r1.RestaurantID.String(), m.RestaurantID().String())
		}
	})
}

func TestMenuRepository_LifecycleTransitions(t *testing.T) {
	repo := NewMenuRepository(testDB, NewOutboxRepository(testDB))
	restRepo := NewRestaurantRepository(testDB, NewOutboxRepository(testDB))

	t.Run("should persist events for menu transitions", func(t *testing.T) {
		truncateTables(testDB)

		addr, _ := valueobjects.NewAddress("St", "1", "", "Dst", "City", "ST", "00000-000")
		r, _ := restaurant.NewRestaurant("Menu Lifecycle Test", addr)
		_ = restRepo.Save(context.Background(), r)

		m, _ := menu.NewMenu("Lifecycle Menu", r.RestaurantID)
		cat, _ := menu.NewCategory("Main")
		_ = m.AddCategory(*cat)

		item, _ := menu.NewItemMenu("Burger", "Good", common.NewMoneyFromCents(1000))
		_ = m.AddItemToCategory(cat.CategoryID, *item)

		err := repo.Save(context.Background(), m)
		require.NoError(t, err)
		assert.Equal(t, 3, CountEventsInOutbox(t, testDB, m.ID().String()))

		err = m.Activate()
		require.NoError(t, err)
		err = repo.Save(context.Background(), m)
		require.NoError(t, err)
		assert.Equal(t, 4, CountEventsInOutbox(t, testDB, m.ID().String()))

		newPrice := common.NewMoneyFromCents(1500)
		err = m.UpdateItemPrice(cat.CategoryID, item.ItemID, newPrice)
		require.NoError(t, err)
		err = repo.Save(context.Background(), m)
		require.NoError(t, err)
		assert.Equal(t, 5, CountEventsInOutbox(t, testDB, m.ID().String()))

		err = m.UpdateItemAvailability(cat.CategoryID, item.ItemID, enums.ItemUnavailable)
		require.NoError(t, err)
		err = repo.Save(context.Background(), m)
		require.NoError(t, err)
		assert.Equal(t, 6, CountEventsInOutbox(t, testDB, m.ID().String()))

		payload := GetLastEventPayload(t, testDB, m.ID().String())
		assert.Equal(t, "UNAVAILABLE", payload["new_status"])
	})
}
