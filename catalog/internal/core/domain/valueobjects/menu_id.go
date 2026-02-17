package valueobjects

import (
	"github.com/google/uuid"
	common "github.com/vterry/food-ordering/common/pkg"
)

type MenuID struct {
	common.BaseID[uuid.UUID]
}

func NewMenuID() MenuID {
	return MenuID{
		BaseID: common.NewBaseID(uuid.New()),
	}
}

func (m MenuID) String() string {
	return m.ID().String()
}

func ParseMenuId(menuId string) (MenuID, error) {
	id, err := uuid.Parse(menuId)
	if err != nil {
		return MenuID{}, err
	}
	return MenuID{
		BaseID: common.NewBaseID(id),
	}, nil
}
