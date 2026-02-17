package input

import "context"

type RestaurantService interface {
	CreateRestaurant(ctx context.Context, req CreateRestaurantRequest) (*RestaurantResponse, error)
	GetRestaurant(ctx context.Context, restIdstr string) (*RestaurantResponse, error)
	OpenRestaurant(ctx context.Context, restIdstr string) error
	CloseRestaurant(ctx context.Context, restIdstr string) error
}
