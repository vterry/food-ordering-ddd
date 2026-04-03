package app

import (
	"context"
	"fmt"

	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/input"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
	common "github.com/vterry/food-ordering/common/pkg"
)

var _ input.CatalogQueryService = (*CatalogQueryAppService)(nil)

type CatalogQueryAppService struct {
	queryRepo output.CatalogQueryRepository
}

func NewCatalogQueryAppService(queryRepo output.CatalogQueryRepository) *CatalogQueryAppService {
	return &CatalogQueryAppService{
		queryRepo: queryRepo,
	}
}

func (c *CatalogQueryAppService) GetActiveMenu(ctx context.Context, restaurantId string) (*input.MenuResponse, error) {
	data, err := c.queryRepo.FindActiveMenuRows(ctx, restaurantId)

	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, common.NewNotFoundErr(output.ErrEntityNotFound)
	}

	menu := &input.MenuResponse{
		ID:           data[0].MenuID,
		Name:         data[0].MenuName,
		RestaurantID: restaurantId,
		Status:       data[0].MenuStatus,
		Categories:   make([]*input.CategoryResponse, 0),
	}

	categoryMap := make(map[string]*input.CategoryResponse)

	for _, row := range data {
		if !row.CategoryID.Valid {
			continue
		}

		cat, exists := categoryMap[row.CategoryID.String]
		if !exists {
			cat = &input.CategoryResponse{
				ID:    row.CategoryID.String,
				Name:  row.CategoryName.String,
				Items: make([]*input.ItemResponse, 0),
			}
			categoryMap[row.CategoryID.String] = cat
			menu.Categories = append(menu.Categories, cat)
		}
		if !row.ItemID.Valid {
			continue
		}
		item := &input.ItemResponse{
			ID:          row.ItemID.String,
			Name:        row.ItemName.String,
			Description: row.ItemDescription.String,
			PriceCents:  row.ItemPriceCents.Int64,
			Status:      row.ItemStatus.String,
		}
		cat.Items = append(cat.Items, item)
	}
	return menu, nil
}

func (c *CatalogQueryAppService) ValidateOrder(ctx context.Context, req input.ValidateOrderRequest) (*input.ValidateOrderResponse, error) {

	data, err := c.queryRepo.FindOrderValidationData(ctx, req.RestaurantID, req.ItemIDs)

	if err != nil {
		return nil, err
	}

	if data.RestaurantStatus != enums.RestaurantOpened.String() || !data.HasActiveMenu {
		return &input.ValidateOrderResponse{
			Valid:            false,
			ValidationErrors: []string{"restaurant cannot accept orders"},
		}, nil
	}

	foundItems := make(map[string]output.OrderValidationItem, len(data.Items))
	for _, item := range data.Items {
		foundItems[item.ItemUUID] = item
	}

	var errors []string
	var snapshot []input.ItemSnapshot

	for _, requestID := range req.ItemIDs {
		item, found := foundItems[requestID]
		if !found {
			errors = append(errors, fmt.Sprintf("item %s not found in active menu", requestID))
			continue
		}

		if item.ItemStatus != enums.ItemAvailable.String() {
			errors = append(errors, fmt.Sprintf("item %s is unavailable", requestID))
			continue
		}

		snapshot = append(snapshot, input.ItemSnapshot{
			ID:         item.ItemUUID,
			Name:       item.ItemName,
			PriceCents: item.PriceCents,
		})
	}

	if len(errors) > 0 {
		return &input.ValidateOrderResponse{Valid: false, ValidationErrors: errors}, nil
	}
	return &input.ValidateOrderResponse{Valid: true, Items: snapshot}, nil
}
