package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
)

func TestRestaurantRepository_Save(t *testing.T) {
	repo := NewRestaurantRepository(testDB)

	tests := []struct {
		name       string
		setup      func(t *testing.T) *restaurant.Restaurant
		wantErr    bool
		assertFunc func(t *testing.T, id valueobjects.RestaurantID)
	}{
		{
			name: "save new restaurant successfully",
			setup: func(t *testing.T) *restaurant.Restaurant {
				addr, err := valueobjects.NewAddress("Test Street", "123", "Apt 101", "District", "City", "ST", "12345-678")
				require.NoError(t, err)
				r, err := restaurant.NewRestaurant("Rest Test", addr)
				require.NoError(t, err)
				return r
			},
			wantErr: false,
			assertFunc: func(t *testing.T, id valueobjects.RestaurantID) {
				found, err := repo.FindById(context.Background(), id)
				require.NoError(t, err)
				assert.Equal(t, "Rest Test", found.Name())

				addr := found.Address()
				assert.Equal(t, "Test Street", addr.Street())

				count := CountEventsInOutbox(t, testDB, id.String())
				assert.Equal(t, 1, count)
				payload := GetLastEventPayload(t, testDB, id.String())
				assert.Equal(t, "Rest Test", payload["name"])
			},
		},
		{
			name: "update existing restaurant",
			setup: func(t *testing.T) *restaurant.Restaurant {
				addr, err := valueobjects.NewAddress("Test Street", "123", "Apt 101", "District", "City", "ST", "12345-678")
				require.NoError(t, err)
				r, err := restaurant.NewRestaurant("Rest Orig", addr)
				require.NoError(t, err)

				err = repo.Save(context.Background(), r)
				require.NoError(t, err)
				return r
			},
			wantErr: false,
			assertFunc: func(t *testing.T, id valueobjects.RestaurantID) {
				found, err := repo.FindById(context.Background(), id)
				require.NoError(t, err)
				assert.Equal(t, "Rest Orig", found.Name())

				count := CountEventsInOutbox(t, testDB, id.String())
				assert.GreaterOrEqual(t, count, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateTables(testDB)
			r := tt.setup(t)
			if r == nil {
				t.Skip("Implement setup")
			}

			err := repo.Save(context.Background(), r)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tt.assertFunc != nil {
				tt.assertFunc(t, r.RestaurantID)
			}
		})
	}
}

func TestRestaurantRepository_FindById(t *testing.T) {
	repo := NewRestaurantRepository(testDB)

	tests := []struct {
		name       string
		seed       func(t *testing.T) valueobjects.RestaurantID
		wantErr    bool
		assertFunc func(t *testing.T, found *restaurant.Restaurant)
	}{
		{
			name: "restaurant not found",
			seed: func(t *testing.T) valueobjects.RestaurantID {
				return valueobjects.NewRestaurantID()
			},
			wantErr: true,
		},
		{
			name: "restaurant found",
			seed: func(t *testing.T) valueobjects.RestaurantID {
				addr, err := valueobjects.NewAddress("Test Street", "123", "Apt 101", "District", "City", "ST", "12345-678")
				require.NoError(t, err)
				r, err := restaurant.NewRestaurant("Rest Find", addr)
				require.NoError(t, err)

				err = repo.Save(context.Background(), r)
				require.NoError(t, err)
				return r.RestaurantID
			},
			wantErr: false,
			assertFunc: func(t *testing.T, found *restaurant.Restaurant) {
				assert.NotNil(t, found)
				assert.Equal(t, "Rest Find", found.Name())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			truncateTables(testDB)
			id := tt.seed(t)
			if id.IsZero() && tt.name != "restaurant not found" {
				t.Skip("Implement seed")
			}

			found, err := repo.FindById(context.Background(), id)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tt.assertFunc != nil {
				tt.assertFunc(t, found)
			}
		})
	}
}

func TestRestaurantRepository_FindAll(t *testing.T) {
	repo := NewRestaurantRepository(testDB)

	t.Run("find all returns all restaurants", func(t *testing.T) {
		truncateTables(testDB)

		addr, _ := valueobjects.NewAddress("St A", "1", "", "Dst B", "City C", "ST", "00000-000")
		r1, err := restaurant.NewRestaurant("Rest 1", addr)
		require.NoError(t, err)
		r2, err := restaurant.NewRestaurant("Rest 2", addr)
		require.NoError(t, err)

		err = repo.Save(context.Background(), r1)
		require.NoError(t, err)
		err = repo.Save(context.Background(), r2)
		require.NoError(t, err)

		all, err := repo.FindAll(context.Background())
		require.NoError(t, err)

		assert.NotEmpty(t, all)
		assert.Len(t, all, 2)
	})
}

func TestRestaurantRepository_StateTransitions(t *testing.T) {
	repo := NewRestaurantRepository(testDB)
	menuId := valueobjects.NewMenuID()

	t.Run("should persist events for opening and closing", func(t *testing.T) {
		truncateTables(testDB)

		addr, _ := valueobjects.NewAddress("St", "1", "", "Dst", "City", "ST", "00000-000")
		r, _ := restaurant.NewRestaurant("State Test", addr)

		err := repo.Save(context.Background(), r)
		require.NoError(t, err)
		assert.Equal(t, 1, CountEventsInOutbox(t, testDB, r.ID().String()))

		err = r.UpdateMenu(menuId)
		require.NoError(t, err)
		err = repo.Save(context.Background(), r)
		require.NoError(t, err)
		assert.Equal(t, 2, CountEventsInOutbox(t, testDB, r.ID().String()))

		payload := GetLastEventPayload(t, testDB, r.ID().String())
		assert.Equal(t, menuId.String(), payload["menu_id"])

		err = r.Open()
		require.NoError(t, err)
		err = repo.Save(context.Background(), r)
		require.NoError(t, err)
		assert.Equal(t, 3, CountEventsInOutbox(t, testDB, r.ID().String()))

		err = r.Close()
		require.NoError(t, err)
		err = repo.Save(context.Background(), r)
		require.NoError(t, err)
		assert.Equal(t, 4, CountEventsInOutbox(t, testDB, r.ID().String()))
	})
}
