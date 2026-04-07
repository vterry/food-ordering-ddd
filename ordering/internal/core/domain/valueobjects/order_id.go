package valueobjects

import (
	"github.com/google/uuid"
	common "github.com/vterry/food-ordering/common/pkg"
)

type OrderID struct {
	common.BaseID[uuid.UUID]
}

func NewOrderID() OrderID {
	return OrderID{
		BaseID: common.NewBaseID(uuid.New()),
	}
}

func (o OrderID) String() string {
	return o.ID().String()
}

func ParseOrderID(orderId string) (OrderID, error) {
	id, err := uuid.Parse(orderId)
	if err != nil {
		return OrderID{}, common.NewValidationErr(err)
	}
	return OrderID{
		BaseID: common.NewBaseID(id),
	}, nil
}
