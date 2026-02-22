package app

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/menu"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
	common "github.com/vterry/food-ordering/common/pkg"
)

type mockUoW struct {
	MockRun func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockUoW) Run(ctx context.Context, fn func(ctxContext context.Context) error) error {
	return m.MockRun(ctx, fn)
}

type mockMenuRepo struct {
	output.MenuRepository
	findByIdMock                     func(ctx context.Context, id valueobjects.MenuID) (*menu.Menu, error)
	findActiveMenuByRestaurantIdMock func(ctx context.Context, id valueobjects.RestaurantID) (*menu.Menu, error)
	saveMock                         func(ctx context.Context, m *menu.Menu) error
}

func (m *mockMenuRepo) FindById(ctx context.Context, id valueobjects.MenuID) (*menu.Menu, error) {
	return m.findByIdMock(ctx, id)
}

func (m *mockMenuRepo) FindActiveMenuByRestaurantId(ctx context.Context, id valueobjects.RestaurantID) (*menu.Menu, error) {
	return m.findActiveMenuByRestaurantIdMock(ctx, id)
}

func (m *mockMenuRepo) Save(ctx context.Context, agg *menu.Menu) error {
	return m.saveMock(ctx, agg)
}

type mockRestaurantRepo struct {
	output.RestaurantRepository
	findByIdMock func(ctx context.Context, id valueobjects.RestaurantID) (*restaurant.Restaurant, error)
	saveMock     func(ctx context.Context, r *restaurant.Restaurant) error
}

func (r *mockRestaurantRepo) FindById(ctx context.Context, id valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
	return r.findByIdMock(ctx, id)
}

func (r *mockRestaurantRepo) Save(ctx context.Context, agg *restaurant.Restaurant) error {
	return r.saveMock(ctx, agg)
}

type mockAssigner struct {
	assignMock func(r *restaurant.Restaurant, m *menu.Menu) error
}

func (a *mockAssigner) AssignMenuToRestaurant(restAgg *restaurant.Restaurant, menuAgg *menu.Menu) error {
	return a.assignMock(restAgg, menuAgg)
}

func TestMenuAppService_ActiveMenu(t *testing.T) {
	validContext := context.Background()

	money := common.NewMoneyFromCents(1000)
	item := menu.RestoreItemMenu(valueobjects.ItemID{}, "Item 1", "Desc", money, enums.ItemAvailable)

	catID, _ := valueobjects.ParseCategoryId(uuid.New().String())
	category := menu.RestoreCategory(catID, "Categoria 1", []menu.ItemMenu{*item})

	restID := valueobjects.NewRestaurantID()

	menuUUID := valueobjects.NewMenuID()

	menuAgg, _ := menu.NewMenu("Menu Test", restID)
	menuAgg = menu.Restore(menuUUID, "Menu Teste", restID, enums.MenuDraft, []menu.Category{*category})

	addr, _ := valueobjects.NewAddress("Rua A", "123", "", "Bairro", "Cidade", "SP", "00000000")
	restAgg := restaurant.Restore(restID, "Restaurante Test", addr, enums.RestaurantOpened, valueobjects.MenuID{})

	tests := []struct {
		name        string
		menuID      valueobjects.MenuID
		mockSetup   func(mRepo *mockMenuRepo, rRepo *mockRestaurantRepo, uow *mockUoW, assigner *mockAssigner)
		expectError bool
	}{
		{
			name:   "Should active menu and assign to restaurant successfully",
			menuID: menuUUID,
			mockSetup: func(mRepo *mockMenuRepo, rRepo *mockRestaurantRepo, uow *mockUoW, assigner *mockAssigner) {
				uow.MockRun = func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				}

				mRepo.findByIdMock = func(ctx context.Context, id valueobjects.MenuID) (*menu.Menu, error) {
					return menuAgg, nil
				}

				rRepo.findByIdMock = func(ctx context.Context, id valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return restAgg, nil
				}

				assigner.assignMock = func(r *restaurant.Restaurant, m *menu.Menu) error {
					return nil
				}

				mRepo.saveMock = func(_ context.Context, _ *menu.Menu) error { return nil }
				rRepo.saveMock = func(_ context.Context, _ *restaurant.Restaurant) error { return nil }
			},
			expectError: false,
		},
		{
			name:   "Should return error when failing to save restaurant (transaction rollback scenario)",
			menuID: menuUUID,
			mockSetup: func(mRepo *mockMenuRepo, rRepo *mockRestaurantRepo, uow *mockUoW, assigner *mockAssigner) {
				uow.MockRun = func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				}

				mRepo.findByIdMock = func(ctx context.Context, id valueobjects.MenuID) (*menu.Menu, error) {
					return menuAgg, nil
				}

				rRepo.findByIdMock = func(ctx context.Context, id valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return restAgg, nil
				}

				assigner.assignMock = func(r *restaurant.Restaurant, m *menu.Menu) error {
					return nil
				}

				mRepo.saveMock = func(_ context.Context, _ *menu.Menu) error { return nil }
				rRepo.saveMock = func(_ context.Context, _ *restaurant.Restaurant) error { return errors.New("db error") }
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockMenuRepo{}
			rRepo := &mockRestaurantRepo{}
			uow := &mockUoW{}
			assigner := &mockAssigner{}

			if tt.mockSetup != nil {
				tt.mockSetup(mRepo, rRepo, uow, assigner)
			}

			service := NewMenuAppService(assigner, uow, mRepo, rRepo)
			err := service.ActiveMenu(validContext, tt.menuID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}
