package menu

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	common "github.com/vterry/food-ordering/common/pkg"
)

// ============================================================
// ItemMenu Entity - Creation
// ============================================================

func TestNewItemMenu(t *testing.T) {
	validPrice := common.NewMoneyFromCents(500)

	tests := []struct {
		name        string
		itemName    string
		description string
		price       common.Money
		wantErr     error
	}{
		{"empty name returns error", "", "desc", validPrice, ErrItemNameIsEmpty},
		{"short name returns error", "ab", "desc", validPrice, ErrInvalidItemNameSize},
		{"long name returns error", strings.Repeat("a", MaxLengthItemNameSize+1), "desc", validPrice, ErrInvalidItemNameSize},
		{"empty description returns error", "Suco", "", validPrice, ErrItemDescriptionIsEmpty},
		{"zero price returns error", "Suco", "Suco natural", common.Zero, ErrItemPriceInvalid},
		{"negative price returns error", "Suco", "Suco natural", common.NewMoneyFromCents(-100), ErrItemPriceInvalid},
		{"valid item created successfully", "Suco", "Suco natural", validPrice, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := NewItemMenu(tt.itemName, tt.description, tt.price)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, item)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, item)
			assert.Equal(t, tt.itemName, item.Name())
			assert.Equal(t, tt.description, item.Description())
			assert.Equal(t, tt.price, item.BasePrice())
			assert.Equal(t, enums.ItemAvailable, item.Status())
		})
	}
}

func TestNewItemMenu_EmitsItemMenuCreatedEvent(t *testing.T) {
	item, err := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
	require.NoError(t, err)

	events := item.PullEvent()
	assert.Len(t, events, 1)
	assert.IsType(t, ItemMenuCreated{}, events[0])
}

// ============================================================
// ItemMenu Entity - Availability Transitions
// ============================================================

func TestItemMenu_MarkAvailable(t *testing.T) {
	item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
	item.MarkUnavailable()
	item.PullEvent() // limpa eventos

	item.MarkAvailable()

	assert.Equal(t, enums.ItemAvailable, item.Status())
	events := item.PullEvent()
	assert.Len(t, events, 1)
	assert.IsType(t, ItemMenuAvailabilityChanged{}, events[0])
}

func TestItemMenu_MarkUnavailable(t *testing.T) {
	item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
	item.PullEvent()

	item.MarkUnavailable()

	assert.Equal(t, enums.ItemUnavailable, item.Status())
	events := item.PullEvent()
	assert.Len(t, events, 1)
	assert.IsType(t, ItemMenuAvailabilityChanged{}, events[0])
}

func TestItemMenu_MarkTemporarilyUnavailable(t *testing.T) {
	item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
	item.PullEvent()

	item.MarkTemporarilyUnavailable()

	assert.Equal(t, enums.ItemTempUnavailable, item.Status())
	events := item.PullEvent()
	assert.Len(t, events, 1)
	assert.IsType(t, ItemMenuAvailabilityChanged{}, events[0])
}

// ============================================================
// ItemMenu Entity - Update Operations
// ============================================================

func TestItemMenu_UpdatePrice(t *testing.T) {
	t.Run("update price succeeds", func(t *testing.T) {
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
		item.PullEvent()

		newPrice := common.NewMoneyFromCents(750)
		err := item.UpdatePrice(newPrice)

		require.NoError(t, err)
		assert.Equal(t, newPrice, item.BasePrice())

		events := item.PullEvent()
		assert.Len(t, events, 1)
		assert.IsType(t, ItemMenuPriceChanged{}, events[0])
	})

	t.Run("update price with zero returns error", func(t *testing.T) {
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))

		err := item.UpdatePrice(common.Zero)

		assert.ErrorIs(t, err, ErrItemPriceInvalid)
		assert.Equal(t, common.NewMoneyFromCents(500), item.BasePrice()) // price should not change
	})
}

func TestItemMenu_UpdateName(t *testing.T) {
	t.Run("update name succeeds", func(t *testing.T) {
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
		item.PullEvent()

		err := item.UpdateName("Limonada")

		require.NoError(t, err)
		assert.Equal(t, "Limonada", item.Name())

		events := item.PullEvent()
		assert.Len(t, events, 1)
		assert.IsType(t, ItemMenuNameChanged{}, events[0])
	})

	t.Run("update name with empty returns error", func(t *testing.T) {
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))

		err := item.UpdateName("")

		assert.ErrorIs(t, err, ErrItemNameIsEmpty)
		assert.Equal(t, "Suco", item.Name())
	})

	t.Run("update name with short name returns error", func(t *testing.T) {
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))

		err := item.UpdateName("ab")

		assert.ErrorIs(t, err, ErrInvalidItemNameSize)
	})
}

func TestItemMenu_UpdateDescription(t *testing.T) {
	t.Run("update description succeeds", func(t *testing.T) {
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))
		item.PullEvent()

		err := item.UpdateDescription("Suco de laranja natural")

		require.NoError(t, err)
		assert.Equal(t, "Suco de laranja natural", item.Description())

		events := item.PullEvent()
		assert.Len(t, events, 1)
		assert.IsType(t, ItemMenuDescriptionChanged{}, events[0])
	})

	t.Run("update description with empty returns error", func(t *testing.T) {
		item, _ := NewItemMenu("Suco", "Suco natural", common.NewMoneyFromCents(500))

		err := item.UpdateDescription("")

		assert.ErrorIs(t, err, ErrItemDescriptionIsEmpty)
		assert.Equal(t, "Suco natural", item.Description())
	})
}
