package app

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/input"
)

func TestRestaurantAppService_CreateRestaurant(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		req         input.CreateRestaurantRequest
		mockSetup   func(rRepo *mockRestaurantRepo, uow *mockUoW)
		expectError bool
		expectResp  bool
	}{
		{
			name: "should create restaurant successfully",
			req: input.CreateRestaurantRequest{
				Name: "Pizza Place",
				Address: input.AddressRequest{
					Street: "Rua A", Number: "123", Neighborhood: "Centro",
					City: "SP", State: "SP", ZipCode: "00000000",
				},
			},
			mockSetup: func(rRepo *mockRestaurantRepo, uow *mockUoW) {
				uow.MockRun = func(_ context.Context, fn func(context.Context) error) error {
					return fn(context.Background())
				}
				rRepo.saveMock = func(_ context.Context, _ *restaurant.Restaurant) error { return nil }
			},
			expectResp: true,
		},
		{
			name: "should return error when address is invalid",
			req: input.CreateRestaurantRequest{
				Name: "Pizza Place",
				Address: input.AddressRequest{
					Street: "", Number: "", Neighborhood: "",
					City: "", State: "", ZipCode: "",
				},
			},
			mockSetup:   func(rRepo *mockRestaurantRepo, uow *mockUoW) {},
			expectError: true,
		},
		{
			name: "should return error when name is empty",
			req: input.CreateRestaurantRequest{
				Name: "",
				Address: input.AddressRequest{
					Street: "Rua A", Number: "123", Neighborhood: "Centro",
					City: "SP", State: "SP", ZipCode: "00000000",
				},
			},
			mockSetup:   func(rRepo *mockRestaurantRepo, uow *mockUoW) {},
			expectError: true,
		},
		{
			name: "should return error when save fails",
			req: input.CreateRestaurantRequest{
				Name: "Pizza Place",
				Address: input.AddressRequest{
					Street: "Rua A", Number: "123", Neighborhood: "Centro",
					City: "SP", State: "SP", ZipCode: "00000000",
				},
			},
			mockSetup: func(rRepo *mockRestaurantRepo, uow *mockUoW) {
				uow.MockRun = func(_ context.Context, fn func(context.Context) error) error {
					return fn(context.Background())
				}
				rRepo.saveMock = func(_ context.Context, _ *restaurant.Restaurant) error {
					return errors.New("db error")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rRepo := &mockRestaurantRepo{}
			uow := &mockUoW{}
			tt.mockSetup(rRepo, uow)

			service := NewRestaurantAppService(uow, rRepo)
			resp, err := service.CreateRestaurant(ctx, tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "Pizza Place", resp.Name)
				assert.Equal(t, "CLOSED", resp.Status)
			}
		})
	}
}

func TestRestaurantAppService_GetRestaurant(t *testing.T) {
	ctx := context.Background()
	restID := valueobjects.NewRestaurantID()
	addr, _ := valueobjects.NewAddress("Rua A", "123", "", "Bairro", "Cidade", "SP", "00000000")
	restAgg := restaurant.Restore(restID, "Pizza Place", addr, enums.RestaurantClosed, valueobjects.MenuID{})

	tests := []struct {
		name        string
		mockSetup   func(rRepo *mockRestaurantRepo)
		expectError bool
	}{
		{
			name: "should return restaurant successfully",
			mockSetup: func(rRepo *mockRestaurantRepo) {
				rRepo.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return restAgg, nil
				}
			},
		},
		{
			name: "should return error when not found",
			mockSetup: func(rRepo *mockRestaurantRepo) {
				rRepo.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return nil, errors.New("not found")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rRepo := &mockRestaurantRepo{}
			tt.mockSetup(rRepo)

			service := NewRestaurantAppService(&mockUoW{}, rRepo)
			resp, err := service.GetRestaurant(ctx, restID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "Pizza Place", resp.Name)
			}
		})
	}
}

func TestRestaurantAppService_OpenRestaurant(t *testing.T) {
	ctx := context.Background()
	restID := valueobjects.NewRestaurantID()
	menuID := valueobjects.NewMenuID()
	addr, _ := valueobjects.NewAddress("Rua A", "123", "", "Bairro", "Cidade", "SP", "00000000")

	tests := []struct {
		name        string
		mockSetup   func(rRepo *mockRestaurantRepo, uow *mockUoW)
		expectError bool
	}{
		{
			name: "should open restaurant successfully",
			mockSetup: func(rRepo *mockRestaurantRepo, uow *mockUoW) {
				closedRest := restaurant.Restore(restID, "Pizza Place", addr, enums.RestaurantClosed, menuID)
				rRepo.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return closedRest, nil
				}
				uow.MockRun = func(_ context.Context, fn func(context.Context) error) error {
					return fn(context.Background())
				}
				rRepo.saveMock = func(_ context.Context, _ *restaurant.Restaurant) error { return nil }
			},
		},
		{
			name: "should return error when restaurant not found",
			mockSetup: func(rRepo *mockRestaurantRepo, uow *mockUoW) {
				rRepo.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return nil, errors.New("not found")
				}
			},
			expectError: true,
		},
		{
			name: "should return error when restaurant is already opened",
			mockSetup: func(rRepo *mockRestaurantRepo, uow *mockUoW) {
				openedRest := restaurant.Restore(restID, "Pizza Place", addr, enums.RestaurantOpened, menuID)
				rRepo.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return openedRest, nil
				}
			},
			expectError: true,
		},
		{
			name: "should return error when restaurant has no active menu",
			mockSetup: func(rRepo *mockRestaurantRepo, uow *mockUoW) {
				noMenuRest := restaurant.Restore(restID, "Pizza Place", addr, enums.RestaurantClosed, valueobjects.MenuID{})
				rRepo.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return noMenuRest, nil
				}
			},
			expectError: true,
		},
		{
			name: "should return error when save fails",
			mockSetup: func(rRepo *mockRestaurantRepo, uow *mockUoW) {
				closedRest := restaurant.Restore(restID, "Pizza Place", addr, enums.RestaurantClosed, menuID)
				rRepo.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return closedRest, nil
				}
				uow.MockRun = func(_ context.Context, fn func(context.Context) error) error {
					return fn(context.Background())
				}
				rRepo.saveMock = func(_ context.Context, _ *restaurant.Restaurant) error {
					return errors.New("db error")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rRepo := &mockRestaurantRepo{}
			uow := &mockUoW{}
			tt.mockSetup(rRepo, uow)

			service := NewRestaurantAppService(uow, rRepo)
			err := service.OpenRestaurant(ctx, restID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRestaurantAppService_CloseRestaurant(t *testing.T) {
	ctx := context.Background()
	restID := valueobjects.NewRestaurantID()
	menuID := valueobjects.NewMenuID()
	addr, _ := valueobjects.NewAddress("Rua A", "123", "", "Bairro", "Cidade", "SP", "00000000")

	tests := []struct {
		name        string
		mockSetup   func(rRepo *mockRestaurantRepo, uow *mockUoW)
		expectError bool
	}{
		{
			name: "should close restaurant successfully",
			mockSetup: func(rRepo *mockRestaurantRepo, uow *mockUoW) {
				openedRest := restaurant.Restore(restID, "Pizza Place", addr, enums.RestaurantOpened, menuID)
				rRepo.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return openedRest, nil
				}
				uow.MockRun = func(_ context.Context, fn func(context.Context) error) error {
					return fn(context.Background())
				}
				rRepo.saveMock = func(_ context.Context, _ *restaurant.Restaurant) error { return nil }
			},
		},
		{
			name: "should return error when restaurant not found",
			mockSetup: func(rRepo *mockRestaurantRepo, uow *mockUoW) {
				rRepo.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return nil, errors.New("not found")
				}
			},
			expectError: true,
		},
		{
			name: "should return error when restaurant is already closed",
			mockSetup: func(rRepo *mockRestaurantRepo, uow *mockUoW) {
				closedRest := restaurant.Restore(restID, "Pizza Place", addr, enums.RestaurantClosed, menuID)
				rRepo.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return closedRest, nil
				}
			},
			expectError: true,
		},
		{
			name: "should return error when save fails",
			mockSetup: func(rRepo *mockRestaurantRepo, uow *mockUoW) {
				openedRest := restaurant.Restore(restID, "Pizza Place", addr, enums.RestaurantOpened, menuID)
				rRepo.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return openedRest, nil
				}
				uow.MockRun = func(_ context.Context, fn func(context.Context) error) error {
					return fn(context.Background())
				}
				rRepo.saveMock = func(_ context.Context, _ *restaurant.Restaurant) error {
					return errors.New("db error")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rRepo := &mockRestaurantRepo{}
			uow := &mockUoW{}
			tt.mockSetup(rRepo, uow)

			service := NewRestaurantAppService(uow, rRepo)
			err := service.CloseRestaurant(ctx, restID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
