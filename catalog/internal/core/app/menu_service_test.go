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
	inputPkg "github.com/vterry/food-ordering/catalog/internal/core/ports/input"
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

func TestMenuAppService_ValidateOrder(t *testing.T) {
	ctx := context.Background()

	restID := valueobjects.NewRestaurantID()
	activeMenuID := valueobjects.NewMenuID()

	addr, _ := valueobjects.NewAddress("Rua A", "123", "", "Bairro", "Cidade", "SP", "00000000")
	closedRest := restaurant.Restore(restID, "Fechado", addr, enums.RestaurantClosed, valueobjects.MenuID{})
	openRestNoMenu := restaurant.Restore(restID, "Aberto sem menu", addr, enums.RestaurantOpened, valueobjects.MenuID{})
	openRest := restaurant.Restore(restID, "Aberto", addr, enums.RestaurantOpened, activeMenuID)

	money := common.NewMoneyFromCents(1000)
	itemID := valueobjects.NewItemID()
	item := menu.RestoreItemMenu(itemID, "Suco", "natural", money, enums.ItemAvailable)
	catID := valueobjects.NewCategoryID()
	cat := menu.RestoreCategory(catID, "Bebidas", []menu.ItemMenu{*item})
	activeMenu := menu.Restore(activeMenuID, "Menu", restID, enums.MenuActive, []menu.Category{*cat})

	tests := []struct {
		name          string
		restaurantID  string
		itemIDs       []string
		setupRestRepo func(r *mockRestaurantRepo)
		setupMenuRepo func(r *mockMenuRepo)
		wantValid     bool
		wantErrors    []string
		wantErr       bool
	}{
		{
			name:         "invalid restaurant UUID returns error",
			restaurantID: "not-a-uuid",
			wantErr:      true,
		},
		{
			name:         "restaurant not found returns error",
			restaurantID: restID.String(),
			setupRestRepo: func(r *mockRestaurantRepo) {
				r.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return nil, output.ErrEntityNotFound
				}
			},
			wantErr: true,
		},
		{
			name:         "closed restaurant returns Valid: false",
			restaurantID: restID.String(),
			setupRestRepo: func(r *mockRestaurantRepo) {
				r.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return closedRest, nil
				}
			},
			wantValid:  false,
			wantErrors: []string{"restaurant is closed or has no active menu"},
		},
		{
			name:         "open restaurant with no active menu assigned returns Valid: false",
			restaurantID: restID.String(),
			setupRestRepo: func(r *mockRestaurantRepo) {
				r.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return openRestNoMenu, nil
				}
			},
			wantValid:  false,
			wantErrors: []string{"restaurant is closed or has no active menu"},
		},
		{
			name:         "open restaurant, active menu found, all items valid",
			restaurantID: restID.String(),
			itemIDs:      []string{itemID.String()},
			setupRestRepo: func(r *mockRestaurantRepo) {
				r.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return openRest, nil
				}
			},
			setupMenuRepo: func(r *mockMenuRepo) {
				r.findActiveMenuByRestaurantIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*menu.Menu, error) {
					return activeMenu, nil
				}
			},
			wantValid:  true,
			wantErrors: nil,
		},
		{
			name:         "open restaurant, item not in menu returns Valid: false",
			restaurantID: restID.String(),
			itemIDs:      []string{valueobjects.NewItemID().String()},
			setupRestRepo: func(r *mockRestaurantRepo) {
				r.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return openRest, nil
				}
			},
			setupMenuRepo: func(r *mockMenuRepo) {
				r.findActiveMenuByRestaurantIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*menu.Menu, error) {
					return activeMenu, nil
				}
			},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rRepo := &mockRestaurantRepo{}
			mRepo := &mockMenuRepo{}
			if tt.setupRestRepo != nil {
				tt.setupRestRepo(rRepo)
			}
			if tt.setupMenuRepo != nil {
				tt.setupMenuRepo(mRepo)
			}

			svc := NewMenuAppService(&mockAssigner{}, &mockUoW{}, mRepo, rRepo)
			resp, err := svc.ValidateOrder(ctx, inputPkg.ValidateOrderRequest{
				RestaurantID: tt.restaurantID,
				ItemIDs:      tt.itemIDs,
			})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantValid, resp.Valid)
			if tt.wantErrors != nil {
				assert.Equal(t, tt.wantErrors, resp.ValidationErrors)
			}
		})
	}
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
