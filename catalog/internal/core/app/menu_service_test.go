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
	"github.com/vterry/food-ordering/catalog/internal/core/ports/input"
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
	item := menu.RestoreItemMenu(valueobjects.NewItemID(), "Item 1", "Desc", money, enums.ItemAvailable)

	catID, _ := valueobjects.ParseCategoryId(uuid.New().String())
	category := menu.RestoreCategory(catID, "Categoria 1", []menu.ItemMenu{*item})

	restID := valueobjects.NewRestaurantID()

	menuUUID := valueobjects.NewMenuID()
	menuAgg := menu.Restore(menuUUID, "Menu Teste", restID, enums.MenuDraft, []menu.Category{*category})

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
		{
			name:   "Should return error when menu is not found",
			menuID: menuUUID,
			mockSetup: func(mRepo *mockMenuRepo, rRepo *mockRestaurantRepo, uow *mockUoW, assigner *mockAssigner) {
				uow.MockRun = func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				}

				mRepo.findByIdMock = func(ctx context.Context, id valueobjects.MenuID) (*menu.Menu, error) {
					return nil, errors.New("menu not found")
				}
			},
			expectError: true,
		},
		{
			name:   "Should return error when restaurant is not found",
			menuID: menuUUID,
			mockSetup: func(mRepo *mockMenuRepo, rRepo *mockRestaurantRepo, uow *mockUoW, assigner *mockAssigner) {
				uow.MockRun = func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				}

				mRepo.findByIdMock = func(ctx context.Context, id valueobjects.MenuID) (*menu.Menu, error) {
					return menuAgg, nil
				}

				rRepo.findByIdMock = func(ctx context.Context, id valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return nil, errors.New("restaurant not found")
				}
			},
			expectError: true,
		},
		{
			name:   "Should return error when menu is already active",
			menuID: menuUUID,
			mockSetup: func(mRepo *mockMenuRepo, rRepo *mockRestaurantRepo, uow *mockUoW, assigner *mockAssigner) {

				activeMenu := menu.Restore(menuUUID, "Menu Teste", restID, enums.MenuActive, []menu.Category{*category})

				mRepo.findByIdMock = func(ctx context.Context, id valueobjects.MenuID) (*menu.Menu, error) {
					return activeMenu, nil
				}

				rRepo.findByIdMock = func(ctx context.Context, id valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return restAgg, nil
				}

				uow.MockRun = func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				}
			},
			expectError: true,
		},
		{
			name:   "Should return error when menu assigment fails",
			menuID: menuUUID,
			mockSetup: func(mRepo *mockMenuRepo, rRepo *mockRestaurantRepo, uow *mockUoW, assigner *mockAssigner) {

				mRepo.findByIdMock = func(ctx context.Context, id valueobjects.MenuID) (*menu.Menu, error) {
					return menuAgg, nil
				}

				rRepo.findByIdMock = func(ctx context.Context, id valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return restAgg, nil
				}

				uow.MockRun = func(ctx context.Context, fn func(ctx context.Context) error) error {
					return fn(ctx)
				}

				assigner.assignMock = func(r *restaurant.Restaurant, m *menu.Menu) error {
					return errors.New("assignment failed")
				}
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

func TestMenuAppService_CreateMenu(t *testing.T) {
	ctx := context.Background()
	restID := valueobjects.NewRestaurantID()
	addr, _ := valueobjects.NewAddress("Rua A", "123", "", "Bairro", "Cidade", "SP", "00000000")
	restAgg := restaurant.Restore(restID, "Restaurante", addr, enums.RestaurantClosed, valueobjects.MenuID{})

	tests := []struct {
		name        string
		req         input.CreateMenuRequest
		mockSetup   func(mRepo *mockMenuRepo, rRepo *mockRestaurantRepo, uow *mockUoW)
		expectError bool
		expectResp  bool
	}{
		{
			name: "should create menu successfully",
			req:  input.CreateMenuRequest{Name: "Menu Principal"},
			mockSetup: func(mRepo *mockMenuRepo, rRepo *mockRestaurantRepo, uow *mockUoW) {
				rRepo.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return restAgg, nil
				}
				uow.MockRun = func(_ context.Context, fn func(context.Context) error) error {
					return fn(context.Background())
				}
				mRepo.saveMock = func(_ context.Context, _ *menu.Menu) error { return nil }
			},
			expectResp: true,
		},
		{
			name: "should return error when restaurant not found",
			req:  input.CreateMenuRequest{Name: "Menu Principal"},
			mockSetup: func(mRepo *mockMenuRepo, rRepo *mockRestaurantRepo, uow *mockUoW) {
				rRepo.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return nil, errors.New("restaurant not found")
				}
			},
			expectError: true,
		},
		{
			name: "should return error when menu name is invalid",
			req:  input.CreateMenuRequest{Name: ""},
			mockSetup: func(mRepo *mockMenuRepo, rRepo *mockRestaurantRepo, uow *mockUoW) {
				rRepo.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return restAgg, nil
				}
			},
			expectError: true,
		},
		{
			name: "should return error when save fails",
			req:  input.CreateMenuRequest{Name: "Menu Principal"},
			mockSetup: func(mRepo *mockMenuRepo, rRepo *mockRestaurantRepo, uow *mockUoW) {
				rRepo.findByIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*restaurant.Restaurant, error) {
					return restAgg, nil
				}
				uow.MockRun = func(_ context.Context, fn func(context.Context) error) error {
					return fn(context.Background())
				}
				mRepo.saveMock = func(_ context.Context, _ *menu.Menu) error {
					return errors.New("db error")
				}
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
			tt.mockSetup(mRepo, rRepo, uow)

			service := NewMenuAppService(assigner, uow, mRepo, rRepo)
			resp, err := service.CreateMenu(ctx, restID, tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.req.Name, resp.Name)
			}
		})
	}
}

func TestMenuAppService_GetMenu(t *testing.T) {
	ctx := context.Background()
	menuID := valueobjects.NewMenuID()
	restID := valueobjects.NewRestaurantID()
	menuAgg := menu.Restore(menuID, "Menu Teste", restID, enums.MenuDraft, []menu.Category{})

	tests := []struct {
		name        string
		mockSetup   func(mRepo *mockMenuRepo)
		expectError bool
	}{
		{
			name: "should return menu successfully",
			mockSetup: func(mRepo *mockMenuRepo) {
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return menuAgg, nil
				}
			},
		},
		{
			name: "should return error when menu not found",
			mockSetup: func(mRepo *mockMenuRepo) {
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return nil, errors.New("not found")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockMenuRepo{}
			tt.mockSetup(mRepo)

			service := NewMenuAppService(&mockAssigner{}, &mockUoW{}, mRepo, &mockRestaurantRepo{})
			resp, err := service.GetMenu(ctx, menuID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "Menu Teste", resp.Name)
			}
		})
	}
}

func TestMenuAppService_GetActiveMenu(t *testing.T) {
	ctx := context.Background()
	restID := valueobjects.NewRestaurantID()
	menuID := valueobjects.NewMenuID()
	menuAgg := menu.Restore(menuID, "Menu Ativo", restID, enums.MenuActive, []menu.Category{})

	tests := []struct {
		name        string
		mockSetup   func(mRepo *mockMenuRepo)
		expectError bool
	}{
		{
			name: "should return active menu successfully",
			mockSetup: func(mRepo *mockMenuRepo) {
				mRepo.findActiveMenuByRestaurantIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*menu.Menu, error) {
					return menuAgg, nil
				}
			},
		},
		{
			name: "should return error when no active menu found",
			mockSetup: func(mRepo *mockMenuRepo) {
				mRepo.findActiveMenuByRestaurantIdMock = func(_ context.Context, _ valueobjects.RestaurantID) (*menu.Menu, error) {
					return nil, errors.New("not found")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockMenuRepo{}
			tt.mockSetup(mRepo)

			service := NewMenuAppService(&mockAssigner{}, &mockUoW{}, mRepo, &mockRestaurantRepo{})
			resp, err := service.GetActiveMenu(ctx, restID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "Menu Ativo", resp.Name)
			}
		})
	}
}

func TestMenuAppService_ArchiveMenu(t *testing.T) {
	ctx := context.Background()
	menuID := valueobjects.NewMenuID()
	restID := valueobjects.NewRestaurantID()

	money := common.NewMoneyFromCents(1000)
	item := menu.RestoreItemMenu(valueobjects.NewItemID(), "Item 1", "Desc", money, enums.ItemAvailable)
	catID, _ := valueobjects.ParseCategoryId(uuid.New().String())
	category := menu.RestoreCategory(catID, "Cat 1", []menu.ItemMenu{*item})

	activeMenu := menu.Restore(menuID, "Menu Ativo", restID, enums.MenuActive, []menu.Category{*category})

	tests := []struct {
		name        string
		mockSetup   func(mRepo *mockMenuRepo, uow *mockUoW)
		expectError bool
	}{
		{
			name: "should archive menu successfully",
			mockSetup: func(mRepo *mockMenuRepo, uow *mockUoW) {
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return activeMenu, nil
				}
				uow.MockRun = func(_ context.Context, fn func(context.Context) error) error {
					return fn(context.Background())
				}
				mRepo.saveMock = func(_ context.Context, _ *menu.Menu) error { return nil }
			},
		},
		{
			name: "should return error when menu not found",
			mockSetup: func(mRepo *mockMenuRepo, uow *mockUoW) {
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return nil, errors.New("not found")
				}
			},
			expectError: true,
		},
		{
			name: "should return error when menu is already archived",
			mockSetup: func(mRepo *mockMenuRepo, uow *mockUoW) {
				archivedMenu := menu.Restore(menuID, "Menu", restID, enums.MenuArchived, []menu.Category{})
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return archivedMenu, nil
				}
			},
			expectError: true,
		},
		{
			name: "should return error when save fails",
			mockSetup: func(mRepo *mockMenuRepo, uow *mockUoW) {
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return activeMenu, nil
				}
				uow.MockRun = func(_ context.Context, fn func(context.Context) error) error {
					return fn(context.Background())
				}
				mRepo.saveMock = func(_ context.Context, _ *menu.Menu) error {
					return errors.New("db error")
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockMenuRepo{}
			uow := &mockUoW{}
			tt.mockSetup(mRepo, uow)

			service := NewMenuAppService(&mockAssigner{}, uow, mRepo, &mockRestaurantRepo{})
			err := service.ArchiveMenu(ctx, menuID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMenuAppService_AddCategory(t *testing.T) {
	ctx := context.Background()
	menuID := valueobjects.NewMenuID()
	restID := valueobjects.NewRestaurantID()
	draftMenu := menu.Restore(menuID, "Menu Draft", restID, enums.MenuDraft, []menu.Category{})

	tests := []struct {
		name        string
		req         input.AddCategoryRequest
		mockSetup   func(mRepo *mockMenuRepo, uow *mockUoW)
		expectError bool
	}{
		{
			name: "should add category successfully",
			req:  input.AddCategoryRequest{Name: "Pizzas"},
			mockSetup: func(mRepo *mockMenuRepo, uow *mockUoW) {
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return draftMenu, nil
				}
				uow.MockRun = func(_ context.Context, fn func(context.Context) error) error {
					return fn(context.Background())
				}
				mRepo.saveMock = func(_ context.Context, _ *menu.Menu) error { return nil }
			},
		},
		{
			name: "should return error when menu not found",
			req:  input.AddCategoryRequest{Name: "Pizzas"},
			mockSetup: func(mRepo *mockMenuRepo, uow *mockUoW) {
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return nil, errors.New("not found")
				}
			},
			expectError: true,
		},
		{
			name: "should return error when category name is empty",
			req:  input.AddCategoryRequest{Name: ""},
			mockSetup: func(mRepo *mockMenuRepo, uow *mockUoW) {
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return draftMenu, nil
				}
			},
			expectError: true,
		},
		{
			name: "should return error when menu is not in draft",
			req:  input.AddCategoryRequest{Name: "Pizzas"},
			mockSetup: func(mRepo *mockMenuRepo, uow *mockUoW) {
				activeMenu := menu.Restore(menuID, "Menu", restID, enums.MenuActive, []menu.Category{})
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return activeMenu, nil
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockMenuRepo{}
			uow := &mockUoW{}
			tt.mockSetup(mRepo, uow)

			service := NewMenuAppService(&mockAssigner{}, uow, mRepo, &mockRestaurantRepo{})
			err := service.AddCategory(ctx, menuID, tt.req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMenuAppService_AddItemToCategory(t *testing.T) {
	ctx := context.Background()
	menuID := valueobjects.NewMenuID()
	restID := valueobjects.NewRestaurantID()
	catID, _ := valueobjects.ParseCategoryId(uuid.New().String())
	category := menu.RestoreCategory(catID, "Pizzas", []menu.ItemMenu{})
	draftMenu := menu.Restore(menuID, "Menu Draft", restID, enums.MenuDraft, []menu.Category{*category})

	tests := []struct {
		name        string
		catID       valueobjects.CategoryID
		req         input.AddItemRequest
		mockSetup   func(mRepo *mockMenuRepo, uow *mockUoW)
		expectError bool
	}{
		{
			name:  "should add item to category successfully",
			catID: catID,
			req:   input.AddItemRequest{Name: "Margherita", Description: "Classic", PriceCents: 2500},
			mockSetup: func(mRepo *mockMenuRepo, uow *mockUoW) {
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return draftMenu, nil
				}
				uow.MockRun = func(_ context.Context, fn func(context.Context) error) error {
					return fn(context.Background())
				}
				mRepo.saveMock = func(_ context.Context, _ *menu.Menu) error { return nil }
			},
		},
		{
			name:  "should return error when menu not found",
			catID: catID,
			req:   input.AddItemRequest{Name: "Margherita", Description: "Classic", PriceCents: 2500},
			mockSetup: func(mRepo *mockMenuRepo, uow *mockUoW) {
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return nil, errors.New("not found")
				}
			},
			expectError: true,
		},
		{
			name:  "should return error when item name is empty",
			catID: catID,
			req:   input.AddItemRequest{Name: "", Description: "Desc", PriceCents: 2500},
			mockSetup: func(mRepo *mockMenuRepo, uow *mockUoW) {
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return draftMenu, nil
				}
			},
			expectError: true,
		},
		{
			name:  "should return error when menu is not in draft",
			catID: catID,
			req:   input.AddItemRequest{Name: "Margherita", Description: "Classic", PriceCents: 2500},
			mockSetup: func(mRepo *mockMenuRepo, uow *mockUoW) {
				activeMenu := menu.Restore(menuID, "Menu", restID, enums.MenuActive, []menu.Category{*category})
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return activeMenu, nil
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockMenuRepo{}
			uow := &mockUoW{}
			tt.mockSetup(mRepo, uow)

			service := NewMenuAppService(&mockAssigner{}, uow, mRepo, &mockRestaurantRepo{})
			err := service.AddItemToCategory(ctx, menuID, tt.catID, tt.req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMenuAppService_UpdateItem(t *testing.T) {
	ctx := context.Background()
	menuID := valueobjects.NewMenuID()
	restID := valueobjects.NewRestaurantID()

	money := common.NewMoneyFromCents(2500)
	itemID := valueobjects.NewItemID()
	item := menu.RestoreItemMenu(itemID, "Pizza", "Classic", money, enums.ItemAvailable)
	catID, _ := valueobjects.ParseCategoryId(uuid.New().String())
	category := menu.RestoreCategory(catID, "Pizzas", []menu.ItemMenu{*item})
	activeMenu := menu.Restore(menuID, "Menu Ativo", restID, enums.MenuActive, []menu.Category{*category})

	tests := []struct {
		name        string
		req         input.UpdateItemRequest
		mockSetup   func(mRepo *mockMenuRepo, uow *mockUoW)
		expectError bool
	}{
		{
			name: "should update item successfully",
			req:  input.UpdateItemRequest{PriceCents: 3000, Status: "AVAILABLE"},
			mockSetup: func(mRepo *mockMenuRepo, uow *mockUoW) {
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return activeMenu, nil
				}
				uow.MockRun = func(_ context.Context, fn func(context.Context) error) error {
					return fn(context.Background())
				}
				mRepo.saveMock = func(_ context.Context, _ *menu.Menu) error { return nil }
			},
		},
		{
			name: "should return error when menu not found",
			req:  input.UpdateItemRequest{PriceCents: 3000, Status: "AVAILABLE"},
			mockSetup: func(mRepo *mockMenuRepo, uow *mockUoW) {
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return nil, errors.New("not found")
				}
			},
			expectError: true,
		},
		{
			name: "should return error when status is invalid",
			req:  input.UpdateItemRequest{PriceCents: 3000, Status: "INVALID"},
			mockSetup: func(mRepo *mockMenuRepo, uow *mockUoW) {
				mRepo.findByIdMock = func(_ context.Context, _ valueobjects.MenuID) (*menu.Menu, error) {
					return activeMenu, nil
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mRepo := &mockMenuRepo{}
			uow := &mockUoW{}
			tt.mockSetup(mRepo, uow)

			service := NewMenuAppService(&mockAssigner{}, uow, mRepo, &mockRestaurantRepo{})
			err := service.UpdateItem(ctx, menuID, catID, itemID, tt.req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
