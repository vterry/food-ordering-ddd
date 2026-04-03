package output

import (
	"context"
	"database/sql"
)

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

type ActiveMenuRow struct {
	MenuID          string
	MenuName        string
	MenuStatus      string
	CategoryID      sql.NullString
	CategoryName    sql.NullString
	ItemID          sql.NullString
	ItemName        sql.NullString
	ItemDescription sql.NullString
	ItemPriceCents  sql.NullInt64
	ItemStatus      sql.NullString
}

type CatalogQueryRepository interface {
	FindOrderValidationData(ctx context.Context, restaurantID string, itemsIDs []string) (*OrderValidationData, error)
	FindActiveMenuRows(ctx context.Context, restaurantId string) ([]ActiveMenuRow, error)
}
