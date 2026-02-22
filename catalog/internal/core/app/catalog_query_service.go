package app

import (
	"context"
	"fmt"

	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/input"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
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
