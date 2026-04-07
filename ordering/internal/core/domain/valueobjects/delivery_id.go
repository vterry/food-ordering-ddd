package valueobjects

import (
	"github.com/google/uuid"
	common "github.com/vterry/food-ordering/common/pkg"
)

type DeliveryID struct {
	common.BaseID[uuid.UUID]
}

func NewDeliveryID() DeliveryID {
	return DeliveryID{
		BaseID: common.NewBaseID(uuid.New()),
	}
}

func (d DeliveryID) String() string {
	return d.ID().String()
}

func ParseDeliveryID(id string) (DeliveryID, error) {
	parsedId, err := uuid.Parse(id)
	if err != nil {
		return DeliveryID{}, common.NewValidationErr(err)
	}
	return DeliveryID{
		BaseID: common.NewBaseID(parsedId),
	}, nil
}
