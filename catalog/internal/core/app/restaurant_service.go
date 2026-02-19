package app

import (
	"context"

	"github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/input"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/output"
)

var _ input.RestaurantService = (*RestaurantAppService)(nil)

type RestaurantAppService struct {
	uow                  output.UnitOfWork
	menuRepository       output.MenuRepository
	restaurantRepository output.RestaurantRepository
}

func NewRestaurantAppService(uow output.UnitOfWork, menuRepo output.MenuRepository, restaurantRepo output.RestaurantRepository) *RestaurantAppService {
	return &RestaurantAppService{
		uow:                  uow,
		menuRepository:       menuRepo,
		restaurantRepository: restaurantRepo,
	}
}

func (r *RestaurantAppService) CreateRestaurant(ctx context.Context, req input.CreateRestaurantRequest) (*input.RestaurantResponse, error) {
	addressVO, err := valueobjects.NewAddress(req.Address.Street, req.Address.Number, req.Address.Complement, req.Address.Neighborhood, req.Address.City,
		req.Address.State, req.Address.ZipCode)
	if err != nil {
		return nil, err
	}

	restaurantAgg, err := restaurant.NewRestaurant(req.Name, addressVO)
	if err != nil {
		return nil, err
	}

	err = r.uow.Run(ctx, func(ctxTx context.Context) error {
		return r.restaurantRepository.Save(ctxTx, restaurantAgg)
	})

	if err != nil {
		return nil, err
	}
	return toRestaurantResponse(restaurantAgg), nil
}

func (r *RestaurantAppService) GetRestaurant(ctx context.Context, restId valueobjects.RestaurantID) (*input.RestaurantResponse, error) {
	restaurantAgg, err := r.restaurantRepository.FindById(ctx, restId)
	if err != nil {
		return nil, err
	}
	return toRestaurantResponse(restaurantAgg), nil
}

func (r *RestaurantAppService) OpenRestaurant(ctx context.Context, restId valueobjects.RestaurantID) error {
	restaurantAgg, err := r.restaurantRepository.FindById(ctx, restId)
	if err != nil {
		return err
	}

	if err = restaurantAgg.Open(); err != nil {
		return err
	}

	return r.uow.Run(ctx, func(ctxTx context.Context) error {
		return r.restaurantRepository.Save(ctxTx, restaurantAgg)
	})
}

func (r *RestaurantAppService) CloseRestaurant(ctx context.Context, restId valueobjects.RestaurantID) error {
	restaurantAgg, err := r.restaurantRepository.FindById(ctx, restId)
	if err != nil {
		return err
	}

	if err = restaurantAgg.Close(); err != nil {
		return err
	}

	return r.uow.Run(ctx, func(ctxTx context.Context) error {
		return r.restaurantRepository.Save(ctxTx, restaurantAgg)
	})
}
