package app

import (
	"context"

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
