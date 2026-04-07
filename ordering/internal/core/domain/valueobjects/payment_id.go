package valueobjects

import (
	"github.com/google/uuid"
	common "github.com/vterry/food-ordering/common/pkg"
)

type PaymentID struct {
	common.BaseID[uuid.UUID]
}

func NewPaymentID() PaymentID {
	return PaymentID{
		BaseID: common.NewBaseID(uuid.New()),
	}
}

func (d PaymentID) String() string {
	return d.ID().String()
}

func ParsePaymentID(id string) (PaymentID, error) {
	parsedId, err := uuid.Parse(id)
	if err != nil {
		return PaymentID{}, common.NewValidationErr(err)
	}
	return PaymentID{
		BaseID: common.NewBaseID(parsedId),
	}, nil
}
