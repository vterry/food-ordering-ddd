package output

import (
	"context"

	"github.com/vterry/food-ordering/catalog/internal/core/domain/menu"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
)

type MenuRepository interface {
	Save(ctx context.Context, menu *menu.Menu) error
	FindById(ctx context.Context, menuId valueobjects.MenuID) (*menu.Menu, error)
	FindByRestaurantId(ctx context.Context, restaurantId valueobjects.RestaurantID) ([]*menu.Menu, error)
	FindActiveMenuByRestaurantId(ctx context.Context, restaurantId valueobjects.RestaurantID) (*menu.Menu, error)
}
