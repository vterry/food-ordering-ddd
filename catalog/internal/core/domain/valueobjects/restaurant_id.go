package valueobjects

import (
	"github.com/google/uuid"
	common "github.com/vterry/food-ordering/common/pkg"
)

type RestaurantID struct {
	common.BaseID[uuid.UUID]
}

func NewRestaurantID() RestaurantID {
	return RestaurantID{
		BaseID: common.NewBaseID(uuid.New()),
	}
}

func (r RestaurantID) String() string {
	return r.ID().String()
}

func ParseRestaurantId(restaurantId string) (RestaurantID, error) {
	id, err := uuid.Parse(restaurantId)
	if err != nil {
		return RestaurantID{}, err
	}
	return RestaurantID{
		BaseID: common.NewBaseID(id),
	}, nil
}
