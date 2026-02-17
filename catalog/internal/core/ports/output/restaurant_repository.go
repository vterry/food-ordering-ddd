package output

import (
	"context"

	"github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
)

type RestaurantRepository interface {
	Save(ctx context.Context, restaurant *restaurant.Restaurant) error
	FindById(ctx context.Context, restaurantId valueobjects.RestaurantID) (*restaurant.Restaurant, error)
	FindAll(ctx context.Context) ([]*restaurant.Restaurant, error)
}
