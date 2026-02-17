package menu

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	common "github.com/vterry/food-ordering/common/pkg"
)

// ============================================================
// Category Entity - Creation
// ============================================================

func TestNewCategory(t *testing.T) {
	tests := []struct {
		name    string
		catName string
		wantErr error
	}{
		{"empty name returns error", "", ErrCategoryNameIsEmpty},
		{"valid category created", "Drinks", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat, err := NewCategory(tt.catName)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, cat)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, cat)
			assert.Equal(t, tt.catName, cat.Name())
			assert.Empty(t, cat.Items())
			assert.True(t, cat.IsEmpty())
		})
	}
}

// ============================================================
// Category Entity - Item Management
// ============================================================

func TestCategory_AddItem(t *testing.T) {
	t.Run("add item succeeds", func(t *testing.T) {
		cat, _ := NewCategory("Drinks")
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))

		_, err := cat.AddItem(*item)

		require.NoError(t, err)
		assert.Len(t, cat.Items(), 1)
		assert.False(t, cat.IsEmpty())
	})

	t.Run("add duplicate item returns error", func(t *testing.T) {
		cat, _ := NewCategory("Drinks")
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
		_, _ = cat.AddItem(*item)

		_, err := cat.AddItem(*item)

		assert.ErrorIs(t, err, ErrItemAlreadyExist)
	})

	t.Run("cannot exceed max items per category", func(t *testing.T) {
		cat, _ := NewCategory("Drinks")

		for i := 0; i < MaxItemPerCategory; i++ {
			item, _ := NewItemMenu("Item", "Desc do item", common.NewMoneyFromCents(100))
			_, _ = cat.AddItem(*item)
		}

		extraItem, _ := NewItemMenu("Extra", "Desc extra", common.NewMoneyFromCents(100))
		_, err := cat.AddItem(*extraItem)

		assert.ErrorIs(t, err, ErrMaxItemsReached)
		assert.Len(t, cat.Items(), MaxItemPerCategory)
	})
}

func TestCategory_RemoveItem(t *testing.T) {
	t.Run("remove existing item succeeds", func(t *testing.T) {
		cat, _ := NewCategory("Drinks")
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
		_, _ = cat.AddItem(*item)

		err := cat.RemoveItem(item.ItemID)

		require.NoError(t, err)
		assert.True(t, cat.IsEmpty())
	})

	t.Run("remove non-existing item returns error", func(t *testing.T) {
		cat, _ := NewCategory("Drinks")

		err := cat.RemoveItem(valueobjects.NewItemID())

		assert.ErrorIs(t, err, ErrItemNotInCategory)
	})
}

func TestCategory_GetItem(t *testing.T) {
	t.Run("get existing item", func(t *testing.T) {
		cat, _ := NewCategory("Drinks")
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
		_, _ = cat.AddItem(*item)

		found, err := cat.GetItem(item.ItemID)

		require.NoError(t, err)
		assert.Equal(t, item.Name(), found.Name())
	})

	t.Run("get non-existing item returns error", func(t *testing.T) {
		cat, _ := NewCategory("Drinks")

		_, err := cat.GetItem(valueobjects.NewItemID())

		assert.ErrorIs(t, err, ErrItemNotInCategory)
	})
}

func TestCategory_Rename(t *testing.T) {
	t.Run("rename with valid name succeeds", func(t *testing.T) {
		cat, _ := NewCategory("Drinks")

		err := cat.Rename("Beverages")

		require.NoError(t, err)
		assert.Equal(t, "Beverages", cat.Name())
	})

	t.Run("rename with empty name returns error", func(t *testing.T) {
		cat, _ := NewCategory("Drinks")

		err := cat.Rename("")

		assert.ErrorIs(t, err, ErrCategoryNameIsEmpty)
		assert.Equal(t, "Drinks", cat.Name()) // name should not change
	})
}
