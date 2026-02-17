package app

import (
	"context"
	"fmt"

	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/menu"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/input"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
	common "github.com/vterry/food-ordering/common/pkg"
)

var _ input.MenuService = (*MenuAppService)(nil)

type MenuAssigner interface {
	AssignMenuToRestaurant(r *restaurant.Restaurant, m *menu.Menu) error
}

type MenuAppService struct {
	menuAssigner         MenuAssigner
	uow                  output.UnitOfWork
	menuRepository       output.MenuRepository
	restaurantRepository output.RestaurantRepository
}

func NewMenuAppService(menuAssigner MenuAssigner, uow output.UnitOfWork, menuRepository output.MenuRepository, restaurantRepository output.RestaurantRepository) *MenuAppService {
	return &MenuAppService{
		menuAssigner:         menuAssigner,
		uow:                  uow,
		menuRepository:       menuRepository,
		restaurantRepository: restaurantRepository,
	}
}

func (m *MenuAppService) CreateMenu(ctx context.Context, restIdstr string, req input.CreateMenuRequest) (*input.MenuResponse, error) {
	restId, err := valueobjects.ParseRestaurantId(restIdstr)
	if err != nil {
		return nil, err
	}

	_, err = m.restaurantRepository.FindById(ctx, restId)
	if err != nil {
		return nil, err
	}

	menuAggregate, err := menu.NewMenu(req.Name, restId)
	if err != nil {
		return nil, err
	}

	err = m.uow.Run(ctx, func(ctxTx context.Context) error {
		return m.menuRepository.Save(ctxTx, menuAggregate)
	})

	if err != nil {
		return nil, err
	}

	return toMenuResponse(menuAggregate), nil
}

func (m *MenuAppService) GetMenu(ctx context.Context, menuIdstr string) (*input.MenuResponse, error) {
	menuId, err := valueobjects.ParseMenuId(menuIdstr)
	if err != nil {
		return nil, err
	}

	menu, err := m.menuRepository.FindById(ctx, menuId)
	if err != nil {
		return nil, err
	}

	return toMenuResponse(menu), nil
}

func (m *MenuAppService) GetActiveMenu(ctx context.Context, restaurantIdStr string) (*input.MenuResponse, error) {

	restaurantId, err := valueobjects.ParseRestaurantId(restaurantIdStr)
	if err != nil {
		return nil, err
	}

	menu, err := m.menuRepository.FindActiveMenuByRestaurantId(ctx, restaurantId)
	if err != nil {
		return nil, err
	}

	return toMenuResponse(menu), nil
}

func (m *MenuAppService) ActiveMenu(ctx context.Context, menuIdstr string) error {
	menuId, err := valueobjects.ParseMenuId(menuIdstr)
	if err != nil {
		return err
	}

	menuAggregate, err := m.menuRepository.FindById(ctx, menuId)
	if err != nil {
		return err
	}

	restaurant, err := m.restaurantRepository.FindById(ctx, menuAggregate.RestaurantID())
	if err != nil {
		return err
	}

	return m.uow.Run(ctx, func(ctxTx context.Context) error {
		if err = menuAggregate.Activate(); err != nil {
			return err
		}

		if err := m.menuAssigner.AssignMenuToRestaurant(restaurant, menuAggregate); err != nil {
			return err
		}

		if err := m.menuRepository.Save(ctxTx, menuAggregate); err != nil {
			return err
		}

		if err := m.restaurantRepository.Save(ctxTx, restaurant); err != nil {
			return err
		}

		return nil
	})
}

func (m *MenuAppService) ArchiveMenu(ctx context.Context, menuIdstr string) error {
	menuId, err := valueobjects.ParseMenuId(menuIdstr)
	if err != nil {
		return err
	}

	menuAggregate, err := m.menuRepository.FindById(ctx, menuId)
	if err != nil {
		return err
	}

	if err = menuAggregate.Archive(); err != nil {
		return err
	}

	return m.uow.Run(ctx, func(ctxTx context.Context) error {
		return m.menuRepository.Save(ctxTx, menuAggregate)
	})
}

func (m *MenuAppService) AddCategory(ctx context.Context, menuIdstr string, req input.AddCategoryRequest) error {
	menuId, err := valueobjects.ParseMenuId(menuIdstr)
	if err != nil {
		return err
	}

	menuObj, err := m.menuRepository.FindById(ctx, menuId)
	if err != nil {
		return err
	}

	category, err := menu.NewCategory(req.Name)
	if err != nil {
		return err
	}

	err = menuObj.AddCategory(*category)
	if err != nil {
		return err
	}

	return m.uow.Run(ctx, func(ctxTx context.Context) error {
		return m.menuRepository.Save(ctxTx, menuObj)
	})
}

func (m *MenuAppService) AddItemToCategory(ctx context.Context, menuIdstr, categoryIdstr string, req input.AddItemRequest) error {
	menuId, err := valueobjects.ParseMenuId(menuIdstr)
	if err != nil {
		return err
	}

	categoryId, err := valueobjects.ParseCategoryId(categoryIdstr)
	if err != nil {
		return err
	}

	menuObj, err := m.menuRepository.FindById(ctx, menuId)
	if err != nil {
		return err
	}

	item, err := menu.NewItemMenu(req.Name, req.Description, common.NewMoneyFromCents(req.PriceCents))
	if err != nil {
		return err
	}

	err = menuObj.AddItemToCategory(categoryId, *item)
	if err != nil {
		return err
	}

	return m.uow.Run(ctx, func(ctxTx context.Context) error {
		return m.menuRepository.Save(ctxTx, menuObj)
	})
}

func (m *MenuAppService) UpdateItem(ctx context.Context, menuIdStr, categoryIdStr, itemIdStr string, req input.UpdateItemRequest) error {
	menuId, err := valueobjects.ParseMenuId(menuIdStr)
	if err != nil {
		return err
	}

	categoryId, err := valueobjects.ParseCategoryId(categoryIdStr)
	if err != nil {
		return err
	}

	itemId, err := valueobjects.ParseItemId(itemIdStr)
	if err != nil {
		return err
	}

	menu, err := m.menuRepository.FindById(ctx, menuId)
	if err != nil {
		return err
	}

	itemStatus, err := enums.ParseItemStatus(req.Status)
	if err != nil {
		return err
	}

	if err := menu.UpdateItemPrice(categoryId, itemId, common.NewMoneyFromCents(req.PriceCents)); err != nil {
		return err
	}

	if err := menu.UpdateItemAvailability(categoryId, itemId, itemStatus); err != nil {
		return err
	}

	return m.uow.Run(ctx, func(ctxTx context.Context) error {
		return m.menuRepository.Save(ctxTx, menu)
	})
}

func (m *MenuAppService) ValidateOrder(ctx context.Context, req input.ValidateOrderRequest) (*input.ValidateOrderResponse, error) {
	restId, err := valueobjects.ParseRestaurantId(req.RestaurantID)
	if err != nil {
		return nil, err
	}
	activeMenu, err := m.menuRepository.FindActiveMenuByRestaurantId(ctx, restId)
	if err != nil {
		return nil, err
	}
	if activeMenu == nil {
		return &input.ValidateOrderResponse{
			Valid:            false,
			ValidationErrors: []string{"Restaurant is closed or has no active menu"},
		}, nil
	}
	return m.validateItems(activeMenu, req.ItemIDs)
}

func (m *MenuAppService) validateItems(menu *menu.Menu, itemIDs []string) (*input.ValidateOrderResponse, error) {
	response := &input.ValidateOrderResponse{
		Valid: true,
		Items: make([]input.ItemSnapshot, 0, len(itemIDs)),
	}
	var errors []string
	for _, idStr := range itemIDs {
		if errStr := m.validateSingleItem(menu, idStr, &response.Items); errStr != "" {
			response.Valid = false
			errors = append(errors, errStr)
		}
	}

	response.ValidationErrors = errors
	return response, nil
}

func (m *MenuAppService) validateSingleItem(menu *menu.Menu, idStr string, items *[]input.ItemSnapshot) string {
	itemID, err := valueobjects.ParseItemId(idStr)
	if err != nil {
		return fmt.Sprintf("Invalid item ID format: %s", idStr)
	}
	item, found := menu.FindItem(itemID)
	if !found {
		return fmt.Sprintf("item %s not found in active menu", idStr)
	}
	if item.Status() != enums.ItemAvailable {
		return fmt.Sprintf("item %s is not available", item.Name())
	}

	*items = append(*items, input.ItemSnapshot{
		ID:         item.ItemID.String(),
		Name:       item.Name(),
		PriceCents: item.BasePrice().Amount(),
	})

	return ""
}
