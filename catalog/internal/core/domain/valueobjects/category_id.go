package valueobjects

import (
	"github.com/google/uuid"
	common "github.com/vterry/food-ordering/common/pkg"
)

type CategoryID struct {
	common.BaseID[uuid.UUID]
}

func NewCategoryID() CategoryID {
	return CategoryID{
		BaseID: common.NewBaseID(uuid.New()),
	}
}

func (c CategoryID) String() string {
	return c.ID().String()
}

func ParseCategoryId(catId string) (CategoryID, error) {
	id, err := uuid.Parse(catId)
	if err != nil {
		return CategoryID{}, common.NewValidationErr(err)
	}
	return CategoryID{
		BaseID: common.NewBaseID(id),
	}, nil
}
