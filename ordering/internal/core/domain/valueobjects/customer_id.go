package valueobjects

import (
	"github.com/google/uuid"
	common "github.com/vterry/food-ordering/common/pkg"
)

type CustomerID struct {
	common.BaseID[uuid.UUID]
}

func NewCustomerID() CustomerID {
	return CustomerID{
		BaseID: common.NewBaseID(uuid.New()),
	}
}

func (c CustomerID) String() string {
	return c.ID().String()
}

func ParseCustomerID(id string) (CustomerID, error) {
	parsedId, err := uuid.Parse(id)
	if err != nil {
		return CustomerID{}, common.NewValidationErr(err)
	}
	return CustomerID{
		BaseID: common.NewBaseID(parsedId),
	}, nil
}
