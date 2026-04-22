package ports

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

type MenuItemInfo struct {
	ID             vo.ID
	RestaurantID   vo.ID
	RestaurantName string
	Name           string
	Description    string
	Price          vo.Money
	Category       string
	IsAvailable    bool
}

type RestaurantQueryUseCase interface {
	GetMenuItemInfo(ctx context.Context, restaurantID, itemID vo.ID) (*MenuItemInfo, error)
	ListRestaurantMenuItems(ctx context.Context, restaurantID vo.ID) ([]*MenuItemInfo, error)
}
