package services

import (
	"errors"

	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/menu"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
)

var (
	ErrMenuNotOwnedByRestaurant = errors.New("menu not owned by restaurant")
	ErrMenuNotActive            = errors.New("menu is not active")
)

type MenuAssignmentService struct{}

func NewMenuAssignmentService() *MenuAssignmentService {
	return &MenuAssignmentService{}
}

func (s *MenuAssignmentService) AssignMenuToRestaurant(restaurant *restaurant.Restaurant, menu *menu.Menu) error {
	if !menu.RestaurantID().Equals(restaurant.RestaurantID) {
		return ErrMenuNotOwnedByRestaurant
	}

	if menu.Status() != enums.MenuActive {
		return ErrMenuNotActive
	}
	return restaurant.UpdateMenu(menu.MenuID)
}
