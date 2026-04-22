package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	apperr "github.com/vterry/food-project/common/pkg/errors"
	"github.com/vterry/food-project/restaurant/internal/core/domain/menu"
	"github.com/vterry/food-project/restaurant/internal/core/domain/restaurant"
	"github.com/vterry/food-project/restaurant/internal/core/ports"
)

var (
	ErrRestaurantNotFound = apperr.NewNotFoundError("RESTAURANT_NOT_FOUND", "restaurant not found", nil)
	ErrMenuNotFound       = apperr.NewNotFoundError("MENU_NOT_FOUND", "menu not found", nil)
)

type RestaurantService struct {
	restaurantRepo ports.RestaurantRepository
	menuRepo       ports.MenuRepository
}

func NewRestaurantService(rr ports.RestaurantRepository, mr ports.MenuRepository) *RestaurantService {
	return &RestaurantService{
		restaurantRepo: rr,
		menuRepo:       mr,
	}
}

func (s *RestaurantService) CreateRestaurant(ctx context.Context, cmd ports.CreateRestaurantCommand) (vo.ID, error) {
	id := vo.NewID(uuid.New().String())
	r := restaurant.NewRestaurant(id, cmd.Name, cmd.Address, cmd.Hours)

	if err := s.restaurantRepo.Save(ctx, r); err != nil {
		return vo.ID{}, apperr.NewInfrastructureError("DATABASE_ERROR", "failed to save restaurant", err)
	}

	return id, nil
}

func (s *RestaurantService) GetRestaurant(ctx context.Context, id vo.ID) (*restaurant.Restaurant, error) {
	r, err := s.restaurantRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperr.NewInfrastructureError("DATABASE_ERROR", "failed to find restaurant", err)
	}
	if r == nil {
		return nil, ErrRestaurantNotFound
	}
	return r, nil
}

func (s *RestaurantService) CreateMenu(ctx context.Context, restaurantID vo.ID, name string) (vo.ID, error) {
	id := vo.NewID(uuid.New().String())
	m := menu.NewMenu(id, restaurantID, name)

	if err := s.menuRepo.Save(ctx, m); err != nil {
		return vo.ID{}, apperr.NewInfrastructureError("DATABASE_ERROR", "failed to save menu", err)
	}

	return id, nil
}

func (s *RestaurantService) ActivateMenu(ctx context.Context, menuID vo.ID) error {
	m, err := s.menuRepo.FindByID(ctx, menuID)
	if err != nil {
		return apperr.NewInfrastructureError("DATABASE_ERROR", "failed to find menu", err)
	}
	if m == nil {
		return ErrMenuNotFound
	}

	// Invariant Check: deactivate currently active menu
	active, err := s.menuRepo.FindActiveByRestaurantID(ctx, m.RestaurantID())
	if err != nil {
		return apperr.NewInfrastructureError("DATABASE_ERROR", "failed to find active menu", err)
	}
	if active != nil && !active.ID().Equals(menuID) {
		active.Deactivate()
		if err := s.menuRepo.Save(ctx, active); err != nil {
			return apperr.NewInfrastructureError("DATABASE_ERROR", "failed to deactivate old menu", err)
		}
	}

	m.Activate()
	if err := s.menuRepo.Save(ctx, m); err != nil {
		return apperr.NewInfrastructureError("DATABASE_ERROR", "failed to activate menu", err)
	}
	return nil
}

func (s *RestaurantService) AddItemToMenu(ctx context.Context, cmd ports.AddMenuItemCommand) (vo.ID, error) {
	m, err := s.menuRepo.FindByID(ctx, cmd.MenuID)
	if err != nil {
		return vo.ID{}, apperr.NewInfrastructureError("DATABASE_ERROR", "failed to find menu", err)
	}
	if m == nil {
		return vo.ID{}, ErrMenuNotFound
	}

	itemID := vo.NewID(uuid.New().String())
	item := menu.NewMenuItem(itemID, cmd.Name, cmd.Description, cmd.Price, cmd.Category)
	
	m.AddItem(item)
	
	if err := s.menuRepo.Save(ctx, m); err != nil {
		return vo.ID{}, apperr.NewInfrastructureError("DATABASE_ERROR", "failed to save menu item", err)
	}
	
	return itemID, nil
}

func (s *RestaurantService) UpdateItemAvailability(ctx context.Context, menuID, productID vo.ID, available bool) error {
	m, err := s.menuRepo.FindByID(ctx, menuID)
	if err != nil {
		return apperr.NewInfrastructureError("DATABASE_ERROR", "failed to find menu", err)
	}
	if m == nil {
		return ErrMenuNotFound
	}

	m.ChangeItemAvailability(productID, available)
	if err := s.menuRepo.Save(ctx, m); err != nil {
		return apperr.NewInfrastructureError("DATABASE_ERROR", "failed to update item availability", err)
	}
	return nil
}
