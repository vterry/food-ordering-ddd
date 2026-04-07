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

func ParseRestaurantID(id string) (RestaurantID, error) {
	parsedId, err := uuid.Parse(id)
	if err != nil {
		return RestaurantID{}, common.NewValidationErr(err)
	}
	return RestaurantID{
		BaseID: common.NewBaseID(parsedId),
	}, nil
}
