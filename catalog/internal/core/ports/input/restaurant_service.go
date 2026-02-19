package input

import (
	"context"

	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
)

type RestaurantService interface {
	CreateRestaurant(ctx context.Context, req CreateRestaurantRequest) (*RestaurantResponse, error)
	GetRestaurant(ctx context.Context, restId valueobjects.RestaurantID) (*RestaurantResponse, error)
	OpenRestaurant(ctx context.Context, restId valueobjects.RestaurantID) error
	CloseRestaurant(ctx context.Context, restId valueobjects.RestaurantID) error
}
