package restaurant

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
)

// ============================================================
// Helpers
// ============================================================

func validAddress() valueobjects.Address {
	addr, _ := valueobjects.NewAddress("Rua das Flores", "123", "", "Centro", "São Paulo", "SP", "01000-000")
	return addr
}

// ============================================================
// Restaurant Aggregate - Creation
// ============================================================

func TestNewRestaurant(t *testing.T) {
	addr := validAddress()

	tests := []struct {
		name    string
		resName string
		address valueobjects.Address
		wantErr error
	}{
		{"empty name returns error", "", addr, ErrNameIsEmpty},
		{"short name returns error", "abc", addr, ErrInvalidNameSize},
		{"long name returns error", "1234567890123456789012345678901", addr, ErrInvalidNameSize},
		{"empty street returns error", "ValidName", valueobjects.Address{}, valueobjects.ErrStreetIsEmpty},
		{"valid restaurant created", "ValidName", addr, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRestaurant(tt.resName, tt.address)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, r)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, r)
			assert.Equal(t, tt.resName, r.Name())
			assert.Equal(t, enums.RestaurantClosed, r.Status())
			assert.Equal(t, addr, r.Address())
		})
	}
}

func TestNewRestaurant_EmitsRestaurantCreatedEvent(t *testing.T) {
	r, err := NewRestaurant("ValidName", validAddress())
	require.NoError(t, err)

	events := r.PullEvent()
	assert.Len(t, events, 1)
	assert.IsType(t, RestaurantCreated{}, events[0])

	// Fat event should carry restaurant data
	created := events[0].(RestaurantCreated)
	assert.Equal(t, "ValidName", created.Name)
	assert.Equal(t, "Rua das Flores", created.Street)
	assert.Equal(t, "123", created.Number)
	assert.Equal(t, "01000-000", created.ZipCode)
	assert.Equal(t, "São Paulo", created.City)
	assert.Equal(t, "SP", created.State)
	assert.Equal(t, enums.RestaurantClosed.String(), created.Status)
}

// ============================================================
// Restaurant Aggregate - Open / Close transitions
// ============================================================

func TestRestaurant_Open(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Restaurant
		wantErr error
	}{
		{
			name: "cannot open without active menu",
			setup: func() *Restaurant {
				r, _ := NewRestaurant("MyRestaurant", validAddress())
				return r
			},
			wantErr: ErrNoActiveMenu,
		},
		{
			name: "open with active menu succeeds",
			setup: func() *Restaurant {
				r, _ := NewRestaurant("MyRestaurant", validAddress())
				_ = r.UpdateMenu(valueobjects.NewMenuID())
				return r
			},
			wantErr: nil,
		},
		{
			name: "cannot open already opened restaurant",
			setup: func() *Restaurant {
				r, _ := NewRestaurant("MyRestaurant", validAddress())
				_ = r.UpdateMenu(valueobjects.NewMenuID())
				_ = r.Open()
				return r
			},
			wantErr: ErrAlreadyOpened,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.setup()
			r.PullEvent() // limpa eventos anteriores

			err := r.Open()

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, enums.RestaurantOpened, r.Status())

			events := r.PullEvent()
			assert.Len(t, events, 1)
			assert.IsType(t, RestaurantOpened{}, events[0])
		})
	}
}

func TestRestaurant_Close(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Restaurant
		wantErr error
	}{
		{
			name: "close opened restaurant succeeds",
			setup: func() *Restaurant {
				r, _ := NewRestaurant("MyRestaurant", validAddress())
				_ = r.UpdateMenu(valueobjects.NewMenuID())
				_ = r.Open()
				return r
			},
			wantErr: nil,
		},
		{
			name: "cannot close already closed restaurant",
			setup: func() *Restaurant {
				r, _ := NewRestaurant("MyRestaurant", validAddress())
				return r
			},
			wantErr: ErrAlreadyClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.setup()
			r.PullEvent()

			err := r.Close()

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, enums.RestaurantClosed, r.Status())

			events := r.PullEvent()
			assert.Len(t, events, 1)
			assert.IsType(t, RestaurantClosed{}, events[0])
		})
	}
}

// ============================================================
// Restaurant Aggregate - Menu Management
// ============================================================

func TestRestaurant_UpdateMenu(t *testing.T) {
	t.Run("update menu with valid id succeeds", func(t *testing.T) {
		r, _ := NewRestaurant("MyRestaurant", validAddress())
		menuID := valueobjects.NewMenuID()

		err := r.UpdateMenu(menuID)

		require.NoError(t, err)
		assert.Equal(t, menuID, r.ActiveMenuID())
	})

	t.Run("update menu with nil id returns error", func(t *testing.T) {
		r, _ := NewRestaurant("MyRestaurant", validAddress())

		err := r.UpdateMenu(valueobjects.MenuID{})

		assert.ErrorIs(t, err, ErrMenuIdIsNil)
	})

	t.Run("update menu emits event", func(t *testing.T) {
		r, _ := NewRestaurant("MyRestaurant", validAddress())
		r.PullEvent()
		menuID := valueobjects.NewMenuID()

		_ = r.UpdateMenu(menuID)

		events := r.PullEvent()
		assert.Len(t, events, 1)
		assert.IsType(t, RestaurantMenuUpdated{}, events[0])

		updated := events[0].(RestaurantMenuUpdated)
		assert.Equal(t, menuID.String(), updated.MenuID)
	})
}

// ============================================================
// Restaurant Aggregate - Business Rules
// ============================================================

func TestRestaurant_CanAcceptOrder(t *testing.T) {
	t.Run("opened restaurant with active menu can accept orders", func(t *testing.T) {
		r, _ := NewRestaurant("MyRestaurant", validAddress())
		_ = r.UpdateMenu(valueobjects.NewMenuID())
		_ = r.Open()

		assert.True(t, r.CanAcceptOrder())
	})

	t.Run("closed restaurant cannot accept orders", func(t *testing.T) {
		r, _ := NewRestaurant("MyRestaurant", validAddress())
		_ = r.UpdateMenu(valueobjects.NewMenuID())

		assert.False(t, r.CanAcceptOrder())
	})

	t.Run("opened restaurant without menu cannot accept orders", func(t *testing.T) {
		// Using Restore to simulate a restaurant opened without a menu (edge case)
		r := Restore(
			valueobjects.NewRestaurantID(),
			"MyRestaurant",
			validAddress(),
			enums.RestaurantOpened,
			valueobjects.MenuID{},
		)

		assert.False(t, r.CanAcceptOrder())
	})
}
