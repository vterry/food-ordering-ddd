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

func (m *MenuAppService) CreateMenu(ctx context.Context, restId valueobjects.RestaurantID, req input.CreateMenuRequest) (*input.MenuResponse, error) {

	_, err := m.restaurantRepository.FindById(ctx, restId)
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

func (m *MenuAppService) GetMenu(ctx context.Context, menuId valueobjects.MenuID) (*input.MenuResponse, error) {

	menu, err := m.menuRepository.FindById(ctx, menuId)
	if err != nil {
		return nil, err
	}

	return toMenuResponse(menu), nil
}

func (m *MenuAppService) GetActiveMenu(ctx context.Context, restaurantId valueobjects.RestaurantID) (*input.MenuResponse, error) {

	menu, err := m.menuRepository.FindActiveMenuByRestaurantId(ctx, restaurantId)
	if err != nil {
		return nil, err
	}

	return toMenuResponse(menu), nil
}

func (m *MenuAppService) ActiveMenu(ctx context.Context, menuId valueobjects.MenuID) error {

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

func (m *MenuAppService) ArchiveMenu(ctx context.Context, menuId valueobjects.MenuID) error {

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

func (m *MenuAppService) AddCategory(ctx context.Context, menuId valueobjects.MenuID, req input.AddCategoryRequest) error {

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

func (m *MenuAppService) AddItemToCategory(ctx context.Context, menuId valueobjects.MenuID, categoryId valueobjects.CategoryID, req input.AddItemRequest) error {

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

func (m *MenuAppService) UpdateItem(ctx context.Context, menuId valueobjects.MenuID, categoryId valueobjects.CategoryID, itemId valueobjects.ItemID, req input.UpdateItemRequest) error {
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

	rest, err := m.restaurantRepository.FindById(ctx, restId)
	if err != nil {
		return nil, err
	}

	if !rest.CanAcceptOrder() {
		return &input.ValidateOrderResponse{
			Valid:            false,
			ValidationErrors: []string{"restaurant is closed or has no active menu"},
		}, nil
	}

	activeMenu, err := m.menuRepository.FindActiveMenuByRestaurantId(ctx, restId)
	if err != nil {
		return nil, err
	}
	if activeMenu == nil {
		// defensive: restaurant reports it can accept orders but no active menu found (data inconsistency)
		return &input.ValidateOrderResponse{
			Valid:            false,
			ValidationErrors: []string{"restaurant has no active menu"},
		}, nil
	}

	return m.validateItems(activeMenu, req.ItemIDs)
}

func (m *MenuAppService) validateItems(activeMenu *menu.Menu, itemIDs []string) (*input.ValidateOrderResponse, error) {
	response := &input.ValidateOrderResponse{
		Valid: true,
		Items: make([]input.ItemSnapshot, 0, len(itemIDs)),
	}

	parsedIDs := make([]valueobjects.ItemID, 0, len(itemIDs))
	for _, idStr := range itemIDs {
		id, err := valueobjects.ParseItemId(idStr)
		if err != nil {
			response.Valid = false
			response.ValidationErrors = append(response.ValidationErrors, fmt.Sprintf("invalid item ID format: %s", idStr))
			continue
		}
		parsedIDs = append(parsedIDs, id)
	}

	for _, result := range activeMenu.ValidateItems(parsedIDs) {
		if result.Error != "" {
			response.Valid = false
			response.ValidationErrors = append(response.ValidationErrors, result.Error)
			continue
		}
		response.Items = append(response.Items, input.ItemSnapshot{
			ID:         result.Item.ItemID.String(),
			Name:       result.Item.Name(),
			PriceCents: result.Item.BasePrice().Amount(),
		})
	}

	return response, nil
}
