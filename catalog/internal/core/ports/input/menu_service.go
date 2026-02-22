package input

import (
	"context"

	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
)

type MenuService interface {
	CreateMenu(ctx context.Context, restId valueobjects.RestaurantID, req CreateMenuRequest) (*MenuResponse, error)
	GetMenu(ctx context.Context, menuId valueobjects.MenuID) (*MenuResponse, error)
	GetActiveMenu(ctx context.Context, restId valueobjects.RestaurantID) (*MenuResponse, error)
	ActiveMenu(ctx context.Context, menuId valueobjects.MenuID) error
	ArchiveMenu(ctx context.Context, menuId valueobjects.MenuID) error
	AddCategory(ctx context.Context, menuId valueobjects.MenuID, req AddCategoryRequest) error
	AddItemToCategory(ctx context.Context, menuId valueobjects.MenuID, categoryId valueobjects.CategoryID, req AddItemRequest) error
	UpdateItem(ctx context.Context, menuId valueobjects.MenuID, categoryId valueobjects.CategoryID, itemId valueobjects.ItemID, req UpdateItemRequest) error
}
