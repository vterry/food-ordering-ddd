package menu

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	common "github.com/vterry/food-ordering/common/pkg"
)

// ============================================================
// Menu Aggregate - Creation
// ============================================================

func TestNewMenu(t *testing.T) {
	validRestaurantId := valueobjects.NewRestaurantID()

	tests := []struct {
		name         string
		menuName     string
		restaurantID valueobjects.RestaurantID
		wantErr      error
	}{
		{"empty name returns error", "", validRestaurantId, ErrMenuNameIsEmpty},
		{"short name returns error", "abc", validRestaurantId, ErrInvalidMenuNameSize},
		{"long name returns error", strings.Repeat("a", MaxLengthMenuNameSize+1), validRestaurantId, ErrInvalidMenuNameSize},
		{"nil restaurant id returns error", "ValidMenu", valueobjects.RestaurantID{}, ErrRestaurantIdIsNil},
		{"valid menu created successfully", "ValidMenu", validRestaurantId, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu, err := NewMenu(tt.menuName, tt.restaurantID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, menu)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, menu)
			assert.Equal(t, tt.menuName, menu.Name())
			assert.Equal(t, enums.MenuDraft, menu.Status())
			assert.Empty(t, menu.Categories())
		})
	}
}

func TestNewMenu_EmitsMenuCreatedEvent(t *testing.T) {
	menu, err := NewMenu("ValidMenu", valueobjects.NewRestaurantID())
	require.NoError(t, err)

	events := menu.PullEvent()
	assert.Len(t, events, 1)
	assert.IsType(t, MenuCreated{}, events[0])
}

// ============================================================
// Menu Aggregate - State Transitions (Activate / Archive)
// ============================================================

func TestMenu_Activate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Menu
		wantErr error
	}{
		{
			name: "cannot activate empty menu",
			setup: func() *Menu {
				m, _ := NewMenu("EmptyMenu", valueobjects.NewRestaurantID())
				return m
			},
			wantErr: ErrCannotActivateEmpty,
		},
		{
			name: "activate menu with items succeeds",
			setup: func() *Menu {
				m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
				cat, _ := NewCategory("Drinks")
				item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
				_, _ = cat.AddItem(*item)
				_ = m.AddCategory(*cat)
				return m
			},
			wantErr: nil,
		},
		{
			name: "cannot activate already active menu",
			setup: func() *Menu {
				m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
				cat, _ := NewCategory("Drinks")
				item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
				_, _ = cat.AddItem(*item)
				_ = m.AddCategory(*cat)
				_ = m.Activate()
				return m
			},
			wantErr: ErrAlreadyActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu := tt.setup()
			menu.PullEvent() // limpa eventos anteriores

			err := menu.Activate()

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, enums.MenuActive, menu.Status())

			events := menu.PullEvent()
			assert.Len(t, events, 1)
			assert.IsType(t, MenuActivated{}, events[0])
		})
	}
}

func TestMenu_Archive(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Menu
		wantErr error
	}{
		{
			name: "archive draft menu succeeds",
			setup: func() *Menu {
				m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
				return m
			},
			wantErr: nil,
		},
		{
			name: "archive active menu succeeds",
			setup: func() *Menu {
				m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
				cat, _ := NewCategory("Drinks")
				item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
				_, _ = cat.AddItem(*item)
				_ = m.AddCategory(*cat)
				_ = m.Activate()
				return m
			},
			wantErr: nil,
		},
		{
			name: "cannot archive already archived menu",
			setup: func() *Menu {
				m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
				_ = m.Archive()
				return m
			},
			wantErr: ErrAlreadyArchived,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu := tt.setup()
			menu.PullEvent()

			err := menu.Archive()

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, enums.MenuArchived, menu.Status())

			events := menu.PullEvent()
			assert.Len(t, events, 1)
			assert.IsType(t, MenuArchived{}, events[0])
		})
	}
}

// ============================================================
// Menu Aggregate - Category Management
// ============================================================

func TestMenu_AddCategory(t *testing.T) {
	t.Run("add category to draft menu succeeds", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
		cat, _ := NewCategory("Drinks")

		err := m.AddCategory(*cat)

		require.NoError(t, err)
		assert.Len(t, m.Categories(), 1)
	})

	t.Run("cannot add category to active menu", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
		cat, _ := NewCategory("Drinks")
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
		_, _ = cat.AddItem(*item)
		_ = m.AddCategory(*cat)
		_ = m.Activate()

		newCat, _ := NewCategory("Snacks")
		err := m.AddCategory(*newCat)

		assert.ErrorIs(t, err, ErrMenuNotEditable)
	})

	t.Run("cannot exceed max categories", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())

		for i := 0; i < MaxCategoriesPerMenu; i++ {
			cat, _ := NewCategory("Cat")
			_ = m.AddCategory(*cat)
		}

		extraCat, _ := NewCategory("Extra")
		err := m.AddCategory(*extraCat)

		assert.ErrorIs(t, err, ErrMaxCategoryReached)
		assert.Len(t, m.Categories(), MaxCategoriesPerMenu)
	})
}

func TestMenu_RemoveCategory(t *testing.T) {
	t.Run("remove existing category succeeds", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
		cat, _ := NewCategory("Drinks")
		_ = m.AddCategory(*cat)

		err := m.RemoveCategory(cat.CategoryID)

		require.NoError(t, err)
		assert.Empty(t, m.Categories())
	})

	t.Run("remove non-existing category returns error", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())

		err := m.RemoveCategory(valueobjects.NewCategoryID())

		assert.ErrorIs(t, err, ErrCategoryNotFound)
	})

	t.Run("cannot remove category from active menu", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
		cat, _ := NewCategory("Drinks")
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
		_, _ = cat.AddItem(*item)
		_ = m.AddCategory(*cat)
		_ = m.Activate()

		err := m.RemoveCategory(cat.CategoryID)

		assert.ErrorIs(t, err, ErrMenuNotEditable)
	})
}

func TestMenu_GetCategory(t *testing.T) {
	t.Run("get existing category", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
		cat, _ := NewCategory("Drinks")
		_ = m.AddCategory(*cat)

		found, err := m.GetCategory(cat.CategoryID)

		require.NoError(t, err)
		assert.Equal(t, cat.Name(), found.Name())
	})

	t.Run("get non-existing category returns error", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())

		_, err := m.GetCategory(valueobjects.NewCategoryID())

		assert.ErrorIs(t, err, ErrCategoryNotFound)
	})
}

// ============================================================
// Menu Aggregate - Item Management via Category
// ============================================================

func TestMenu_AddItemToCategory(t *testing.T) {
	t.Run("add item to existing category succeeds", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
		cat, _ := NewCategory("Drinks")
		_ = m.AddCategory(*cat)

		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
		err := m.AddItemToCategory(cat.CategoryID, *item)

		require.NoError(t, err)
		found, _ := m.GetCategory(cat.CategoryID)
		assert.Len(t, found.Items(), 1)
	})

	t.Run("add item to non-existing category returns error", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))

		err := m.AddItemToCategory(valueobjects.NewCategoryID(), *item)

		assert.ErrorIs(t, err, ErrCategoryNotFound)
	})

	t.Run("cannot add item to active menu", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
		cat, _ := NewCategory("Drinks")
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
		_, _ = cat.AddItem(*item)
		_ = m.AddCategory(*cat)
		_ = m.Activate()

		newItem, _ := NewItemMenu("Café", "Café expresso", common.NewMoneyFromCents(300))
		err := m.AddItemToCategory(cat.CategoryID, *newItem)

		assert.ErrorIs(t, err, ErrMenuNotEditable)
	})
}

func TestMenu_RemoveItemFromCategory(t *testing.T) {
	t.Run("remove item from category succeeds", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
		cat, _ := NewCategory("Drinks")
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
		_, _ = cat.AddItem(*item)
		_ = m.AddCategory(*cat)

		err := m.RemoveItemFromCategory(cat.CategoryID, *item)

		require.NoError(t, err)
		found, _ := m.GetCategory(cat.CategoryID)
		assert.Empty(t, found.Items())
	})

	t.Run("cannot remove item from active menu", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
		cat, _ := NewCategory("Drinks")
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
		_, _ = cat.AddItem(*item)
		_ = m.AddCategory(*cat)
		_ = m.Activate()

		err := m.RemoveItemFromCategory(cat.CategoryID, *item)

		assert.ErrorIs(t, err, ErrMenuNotEditable)
	})
}

func TestMenu_UpdateItemPrice(t *testing.T) {
	t.Run("update item price succeeds", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
		cat, _ := NewCategory("Drinks")
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
		_, _ = cat.AddItem(*item)
		_ = m.AddCategory(*cat)

		newPrice := common.NewMoneyFromCents(750)
		err := m.UpdateItemPrice(cat.CategoryID, item.ItemID, newPrice)

		require.NoError(t, err)
		found, _ := m.GetCategory(cat.CategoryID)
		updatedItem, _ := found.GetItem(item.ItemID)
		assert.Equal(t, newPrice, updatedItem.BasePrice())
	})

	t.Run("update item price with zero returns error", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
		cat, _ := NewCategory("Drinks")
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
		_, _ = cat.AddItem(*item)
		_ = m.AddCategory(*cat)

		err := m.UpdateItemPrice(cat.CategoryID, item.ItemID, common.Zero)

		assert.ErrorIs(t, err, ErrItemPriceInvalid)
	})

	t.Run("update item in non-existing category returns error", func(t *testing.T) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())

		err := m.UpdateItemPrice(valueobjects.NewCategoryID(), valueobjects.NewItemID(), common.NewMoneyFromCents(100))

		assert.ErrorIs(t, err, ErrCategoryNotFound)
	})
}

func TestMenu_UpdateItemAvailability(t *testing.T) {
	setup := func() (*Menu, *Category, *ItemMenu) {
		m, _ := NewMenu("MyMenu", valueobjects.NewRestaurantID())
		cat, _ := NewCategory("Drinks")
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
		_, _ = cat.AddItem(*item)
		_ = m.AddCategory(*cat)
		return m, cat, item
	}

	t.Run("mark item as unavailable", func(t *testing.T) {
		m, cat, item := setup()

		err := m.UpdateItemAvailability(cat.CategoryID, item.ItemID, enums.ItemUnavailable)

		require.NoError(t, err)
		found, _ := m.GetCategory(cat.CategoryID)
		updatedItem, _ := found.GetItem(item.ItemID)
		assert.Equal(t, enums.ItemUnavailable, updatedItem.Status())
	})

	t.Run("mark item as temporarily unavailable", func(t *testing.T) {
		m, cat, item := setup()

		err := m.UpdateItemAvailability(cat.CategoryID, item.ItemID, enums.ItemTempUnavailable)

		require.NoError(t, err)
		found, _ := m.GetCategory(cat.CategoryID)
		updatedItem, _ := found.GetItem(item.ItemID)
		assert.Equal(t, enums.ItemTempUnavailable, updatedItem.Status())
	})

	t.Run("mark item as available", func(t *testing.T) {
		m, cat, item := setup()
		_ = m.UpdateItemAvailability(cat.CategoryID, item.ItemID, enums.ItemUnavailable)

		err := m.UpdateItemAvailability(cat.CategoryID, item.ItemID, enums.ItemAvailable)

		require.NoError(t, err)
		found, _ := m.GetCategory(cat.CategoryID)
		updatedItem, _ := found.GetItem(item.ItemID)
		assert.Equal(t, enums.ItemAvailable, updatedItem.Status())
	})

	t.Run("invalid status returns error", func(t *testing.T) {
		m, cat, item := setup()

		err := m.UpdateItemAvailability(cat.CategoryID, item.ItemID, enums.ItemStatus(99))

		assert.ErrorIs(t, err, ErrInvalidItemStatus)
	})
}
