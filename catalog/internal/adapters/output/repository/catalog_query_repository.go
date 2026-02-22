package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
)

var _ output.CatalogQueryRepository = (*CatalogQueryRepository)(nil)

type CatalogQueryRepository struct {
	db *sql.DB
}

func NewCatalogQueryRepository(db *sql.DB) *CatalogQueryRepository {
	return &CatalogQueryRepository{
		db: db,
	}
}

func (c *CatalogQueryRepository) FindOrderValidationData(ctx context.Context, restaurantID string, itemsIDs []string) (*output.OrderValidationData, error) {

	placeHolders := strings.Repeat("?,", len(itemsIDs))
	placeHolders = placeHolders[:len(placeHolders)-1]

	query := fmt.Sprintf(QueryFindOrderValidationData, placeHolders)

	args := make([]any, 0, len(itemsIDs)+1)
	for _, id := range itemsIDs {
		args = append(args, id)
	}

	args = append(args, restaurantID)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result *output.OrderValidationData

	for rows.Next() {
		var restUUID, restStats string
		var hasActiveMenu bool
		var itemUUID, itemName, itemStatus *string
		var priceCents *int64

		rows.Scan(&restUUID, &restStats, &hasActiveMenu, &itemUUID, &itemName, &priceCents, &itemStatus)
		if result == nil {
			result = &output.OrderValidationData{
				RestaurantUUID:   restUUID,
				RestaurantStatus: restStats,
				HasActiveMenu:    hasActiveMenu,
			}
		}

		if itemUUID != nil {
			result.Items = append(result.Items, output.OrderValidationItem{
				ItemUUID:   *itemUUID,
				ItemName:   *itemName,
				PriceCents: *priceCents,
				ItemStatus: *itemStatus,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if result == nil {
		return nil, output.ErrEntityNotFound
	}

	return result, nil
}
