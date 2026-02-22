package output

import "context"

type OrderValidationItem struct {
	ItemUUID   string
	ItemName   string
	PriceCents int64
	ItemStatus string
}

type OrderValidationData struct {
	RestaurantUUID   string
	RestaurantStatus string
	HasActiveMenu    bool
	Items            []OrderValidationItem
}

type CatalogQueryRepository interface {
	FindOrderValidationData(ctx context.Context, restaurantID string, itemsIDs []string) (*OrderValidationData, error)
}
