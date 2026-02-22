package valueobjects

import (
	"github.com/google/uuid"
	common "github.com/vterry/food-ordering/common/pkg"
)

type ItemID struct {
	common.BaseID[uuid.UUID]
}

func NewItemID() ItemID {
	return ItemID{
		BaseID: common.NewBaseID(uuid.New()),
	}
}

func (i ItemID) String() string {
	return i.ID().String()
}

func ParseItemId(itemId string) (ItemID, error) {
	id, err := uuid.Parse(itemId)
	if err != nil {
		return ItemID{}, common.NewValidationErr(err)
	}
	return ItemID{
		BaseID: common.NewBaseID(id),
	}, nil
}
