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
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
	common "github.com/vterry/food-ordering/common/pkg"
)

func TestCatalogQueryRepository_FindOrderValidationData(t *testing.T) {
	queryRepo := NewCatalogQueryRepository(testDB)
	outboxRepo := NewOutboxRepository(testDB)
	restRepo := NewRestaurantRepository(testDB, outboxRepo)
	menuRepo := NewMenuRepository(testDB, outboxRepo)

	ctx := context.Background()

	t.Run("restaurant not found", func(t *testing.T) {
		truncateTables(testDB)
		data, err := queryRepo.FindOrderValidationData(ctx, valueobjects.NewRestaurantID().String(), []string{"item-1"})
		assert.ErrorIs(t, err, output.ErrEntityNotFound)
		assert.Nil(t, data)
	})

	t.Run("restaurant closed", func(t *testing.T) {
		truncateTables(testDB)
		addr, _ := valueobjects.NewAddress("St", "1", "", "Dst", "City", "ST", "00000000")
		r := restaurant.Restore(valueobjects.NewRestaurantID(), "rest 1", addr, enums.RestaurantClosed, valueobjects.MenuID{})
		err := restRepo.Save(ctx, r)
		require.NoError(t, err)

		data, err := queryRepo.FindOrderValidationData(ctx, r.ID().String(), []string{"item-1"})
		require.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, r.ID().String(), data.RestaurantUUID)
		assert.Equal(t, enums.RestaurantClosed.String(), data.RestaurantStatus)
		assert.False(t, data.HasActiveMenu)
		assert.Empty(t, data.Items)
	})

	t.Run("open restaurant without active menu", func(t *testing.T) {
		truncateTables(testDB)
		addr, _ := valueobjects.NewAddress("St", "1", "", "Dst", "City", "ST", "00000000")
		r := restaurant.Restore(valueobjects.NewRestaurantID(), "rest 1", addr, enums.RestaurantOpened, valueobjects.MenuID{})
		err := restRepo.Save(ctx, r)
		require.NoError(t, err)

		data, err := queryRepo.FindOrderValidationData(ctx, r.ID().String(), []string{"item-1"})
		require.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, enums.RestaurantOpened.String(), data.RestaurantStatus)
		assert.False(t, data.HasActiveMenu)
		assert.Empty(t, data.Items)
	})

	t.Run("open restaurant with active menu and valid items", func(t *testing.T) {
		truncateTables(testDB)
		addr, _ := valueobjects.NewAddress("St", "1", "", "Dst", "City", "ST", "00000000")
		r := restaurant.Restore(valueobjects.NewRestaurantID(), "rest 1", addr, enums.RestaurantOpened, valueobjects.MenuID{})
		_ = restRepo.Save(ctx, r)

		m, _ := menu.NewMenu("Menu 1", r.RestaurantID)
		cat, _ := menu.NewCategory("Cat 1")
		_ = m.AddCategory(*cat)

		item1, _ := menu.NewItemMenu("Pizza", "Desc", common.NewMoneyFromCents(2500))
		_ = m.AddItemToCategory(cat.CategoryID, *item1)

		item2, _ := menu.NewItemMenu("Lasagna", "Desc", common.NewMoneyFromCents(3000))
		_ = m.AddItemToCategory(cat.CategoryID, *item2)

		_ = menuRepo.Save(ctx, m)

		_ = m.Activate()
		_ = menuRepo.Save(ctx, m)
		_ = r.UpdateMenu(m.MenuID)
		_ = restRepo.Save(ctx, r)

		data, err := queryRepo.FindOrderValidationData(ctx, r.ID().String(), []string{item1.ItemID.String(), item2.ItemID.String()})
		require.NoError(t, err)

		assert.Equal(t, enums.RestaurantOpened.String(), data.RestaurantStatus)
		assert.True(t, data.HasActiveMenu)
		assert.Len(t, data.Items, 2)

		itemMap := make(map[string]output.OrderValidationItem)
		for _, it := range data.Items {
			itemMap[it.ItemUUID] = it
		}

		assert.Contains(t, itemMap, item1.ItemID.String())
		assert.Equal(t, "Pizza", itemMap[item1.ItemID.String()].ItemName)
		assert.Equal(t, int64(2500), itemMap[item1.ItemID.String()].PriceCents)
		assert.Equal(t, enums.ItemAvailable.String(), itemMap[item1.ItemID.String()].ItemStatus)

		assert.Contains(t, itemMap, item2.ItemID.String())
	})

	t.Run("items not found in menu", func(t *testing.T) {
		truncateTables(testDB)
		addr, _ := valueobjects.NewAddress("St", "1", "", "Dst", "City", "ST", "00000000")
		r := restaurant.Restore(valueobjects.NewRestaurantID(), "rest 1", addr, enums.RestaurantOpened, valueobjects.MenuID{})
		_ = restRepo.Save(ctx, r)

		m, _ := menu.NewMenu("Menu 1", r.RestaurantID)
		cat, _ := menu.NewCategory("Cat 1")
		_ = m.AddCategory(*cat)

		item1, _ := menu.NewItemMenu("Pizza", "Desc", common.NewMoneyFromCents(2500))
		_ = m.AddItemToCategory(cat.CategoryID, *item1)

		_ = menuRepo.Save(ctx, m)
		_ = m.Activate()
		_ = menuRepo.Save(ctx, m)
		_ = r.UpdateMenu(m.MenuID)
		_ = restRepo.Save(ctx, r)

		data, err := queryRepo.FindOrderValidationData(ctx, r.ID().String(), []string{item1.ItemID.String(), "fake-item"})
		require.NoError(t, err)

		assert.True(t, data.HasActiveMenu)
		assert.Len(t, data.Items, 1)
		assert.Equal(t, item1.ItemID.String(), data.Items[0].ItemUUID)
	})

	t.Run("item unavailable", func(t *testing.T) {
		truncateTables(testDB)
		addr, _ := valueobjects.NewAddress("St", "1", "", "Dst", "City", "ST", "00000000")
		r := restaurant.Restore(valueobjects.NewRestaurantID(), "rest 1", addr, enums.RestaurantOpened, valueobjects.MenuID{})
		_ = restRepo.Save(ctx, r)

		m, _ := menu.NewMenu("Menu 1", r.RestaurantID)
		cat, _ := menu.NewCategory("Cat 1")
		_ = m.AddCategory(*cat)

		item1, _ := menu.NewItemMenu("Pizza", "Desc", common.NewMoneyFromCents(2500))
		_ = m.AddItemToCategory(cat.CategoryID, *item1)

		_ = m.Activate()
		_ = menuRepo.Save(ctx, m)
		_ = r.UpdateMenu(m.MenuID)
		_ = restRepo.Save(ctx, r)

		_ = m.UpdateItemAvailability(cat.CategoryID, item1.ItemID, enums.ItemTempUnavailable)
		_ = menuRepo.Save(ctx, m)

		data, err := queryRepo.FindOrderValidationData(ctx, r.ID().String(), []string{item1.ItemID.String()})
		require.NoError(t, err)

		assert.True(t, data.HasActiveMenu)
		assert.Len(t, data.Items, 1)
		assert.Equal(t, enums.ItemTempUnavailable.String(), data.Items[0].ItemStatus)
	})
}
