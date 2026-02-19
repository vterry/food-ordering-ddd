package menu

import (
	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	common "github.com/vterry/food-ordering/common/pkg"
)

type ItemMenu struct {
	valueobjects.ItemID
	name        string
	description string
	basePrice   common.Money
	status      enums.ItemStatus
}

func NewItemMenu(name, description string, price common.Money) (*ItemMenu, error) {

	if err := ValidateNewItemMenu(name, description, price); err != nil {
		return nil, err
	}

	item := &ItemMenu{
		ItemID:      valueobjects.NewItemID(),
		name:        name,
		description: description,
		basePrice:   price,
		status:      enums.ItemAvailable,
	}
	return item, nil
}

func (i *ItemMenu) MarkAvailable() {
	i.status = enums.ItemAvailable
}

func (i *ItemMenu) MarkUnavailable() {
	i.status = enums.ItemUnavailable
}

func (i *ItemMenu) MarkTemporarilyUnavailable() {
	i.status = enums.ItemTempUnavailable
}

func (i *ItemMenu) UpdatePrice(price common.Money) error {

	if err := ValidateItemPrice(price); err != nil {
		return err
	}
	i.basePrice = price

	return nil
}

func (i *ItemMenu) UpdateName(name string) error {
	if err := ValidateItemName(name); err != nil {
		return err
	}

	i.name = name

	return nil
}

func (i *ItemMenu) UpdateDescription(description string) error {

	if err := ValidateItemDescription(description); err != nil {
		return err
	}

	i.description = description

	return nil
}

func (i *ItemMenu) Name() string             { return i.name }
func (i *ItemMenu) Description() string      { return i.description }
func (i *ItemMenu) BasePrice() common.Money  { return i.basePrice }
func (i *ItemMenu) Status() enums.ItemStatus { return i.status }

func RestoreItemMenu(
	itemId valueobjects.ItemID,
	name, description string,
	basePrice common.Money,
	status enums.ItemStatus) *ItemMenu {
	return &ItemMenu{
		ItemID:      itemId,
		name:        name,
		description: description,
		basePrice:   basePrice,
		status:      status,
	}
}
