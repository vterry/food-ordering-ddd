package input

import "context"

type CatalogQueryService interface {
	ValidateOrder(ctx context.Context, req ValidateOrderRequest) (*ValidateOrderResponse, error)
}
