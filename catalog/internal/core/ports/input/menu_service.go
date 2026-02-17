package input

import "context"

type MenuService interface {
	CreateMenu(ctx context.Context, restIdstr string, req CreateMenuRequest) (*MenuResponse, error)
	GetMenu(ctx context.Context, menuIdstr string) (*MenuResponse, error)
	GetActiveMenu(ctx context.Context, restaurantIdStr string) (*MenuResponse, error)
	ActiveMenu(ctx context.Context, menuIdstr string) error
	ArchiveMenu(ctx context.Context, menuIdstr string) error
	AddCategory(ctx context.Context, menuIdstr string, req AddCategoryRequest) error
	AddItemToCategory(ctx context.Context, menuIdstr, categoryIdstr string, req AddItemRequest) error
	UpdateItem(ctx context.Context, menuIdStr, categoryIdStr, itemIdStr string, req UpdateItemRequest) error
	ValidateOrder(ctx context.Context, req ValidateOrderRequest) (*ValidateOrderResponse, error)
}
