package services

import (
	"context"
	"fmt"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/restaurant/internal/core/ports"
)

type RestaurantQueryService struct {
	restaurantRepo ports.RestaurantRepository
	menuRepo       ports.MenuRepository
}

func NewRestaurantQueryService(
	restaurantRepo ports.RestaurantRepository,
	menuRepo ports.MenuRepository,
) *RestaurantQueryService {
	return &RestaurantQueryService{
		restaurantRepo: restaurantRepo,
		menuRepo:       menuRepo,
	}
}

func (s *RestaurantQueryService) GetMenuItemInfo(ctx context.Context, restaurantID, itemID vo.ID) (*ports.MenuItemInfo, error) {
	rest, err := s.restaurantRepo.FindByID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	if rest == nil {
		return nil, fmt.Errorf("restaurant not found")
	}

	menu, err := s.menuRepo.FindActiveByRestaurantID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	if menu == nil {
		return nil, fmt.Errorf("active menu not found")
	}

	for _, item := range menu.Items() {
		if item.ID().String() == itemID.String() {
			return &ports.MenuItemInfo{
				ID:             item.ID(),
				RestaurantID:   restaurantID,
				RestaurantName: rest.Name(),
				Name:           item.Name(),
				Description:    item.Description(),
				Price:          item.Price(),
				IsAvailable:    item.IsAvailable(),
			}, nil
		}
	}

	return nil, fmt.Errorf("menu item not found")
}

func (s *RestaurantQueryService) ListRestaurantMenuItems(ctx context.Context, restaurantID vo.ID) ([]*ports.MenuItemInfo, error) {
	rest, err := s.restaurantRepo.FindByID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	if rest == nil {
		return nil, fmt.Errorf("restaurant not found")
	}

	menu, err := s.menuRepo.FindActiveByRestaurantID(ctx, restaurantID)
	if err != nil {
		return nil, err
	}
	if menu == nil {
		return nil, nil // Or empty slice
	}

	infos := make([]*ports.MenuItemInfo, 0, len(menu.Items()))
	for _, item := range menu.Items() {
		infos = append(infos, &ports.MenuItemInfo{
			ID:             item.ID(),
			RestaurantID:   restaurantID,
			RestaurantName: rest.Name(),
			Name:           item.Name(),
			Description:    item.Description(),
			Price:          item.Price(),
			IsAvailable:    item.IsAvailable(),
		})
	}

	return infos, nil
}
