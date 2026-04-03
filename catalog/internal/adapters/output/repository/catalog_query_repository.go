package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
	common "github.com/vterry/food-ordering/common/pkg"
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

func (c *CatalogQueryRepository) FindActiveMenuRows(ctx context.Context, restaurantId string) ([]output.ActiveMenuRow, error) {
	rows, err := c.db.QueryContext(ctx, QueryCatalogFindActiveMenuRows, restaurantId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []output.ActiveMenuRow
	for rows.Next() {
		var row output.ActiveMenuRow

		err := rows.Scan(
			&row.MenuID, &row.MenuName, &row.MenuStatus,
			&row.CategoryID, &row.CategoryName,
			&row.ItemID, &row.ItemName, &row.ItemDescription, &row.ItemPriceCents, &row.ItemStatus,
		)

		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
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
		return nil, common.NewNotFoundErr(output.ErrEntityNotFound)
	}

	return result, nil
}
