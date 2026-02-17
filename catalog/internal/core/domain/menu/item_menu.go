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
	events      []common.DomainEvent
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

	event := NewItemMenuCreated(item.ItemID, item.name, item.basePrice)
	item.AddEvent(event)

	return item, nil
}

func (i *ItemMenu) MarkAvailable() {
	oldStatus := i.status
	i.status = enums.ItemAvailable
	event := NewItemMenuAvailabilityChanged(i.ItemID, oldStatus, i.status)
	i.AddEvent(event)
}

func (i *ItemMenu) MarkUnavailable() {
	oldStatus := i.status
	i.status = enums.ItemUnavailable
	event := NewItemMenuAvailabilityChanged(i.ItemID, oldStatus, i.status)
	i.AddEvent(event)
}

func (i *ItemMenu) MarkTemporarilyUnavailable() {
	oldStatus := i.status
	i.status = enums.ItemTempUnavailable
	event := NewItemMenuAvailabilityChanged(i.ItemID, oldStatus, i.status)
	i.AddEvent(event)
}

func (i *ItemMenu) UpdatePrice(price common.Money) error {

	if err := ValidateItemPrice(price); err != nil {
		return err
	}
	oldPrice := i.basePrice
	i.basePrice = price

	event := NewItemMenuPriceChanged(i.ItemID, oldPrice, i.basePrice)
	i.AddEvent(event)

	return nil
}

func (i *ItemMenu) UpdateName(name string) error {
	if err := ValidateItemName(name); err != nil {
		return err
	}

	i.name = name
	event := NewItemMenuNameChanged(i.ItemID, name)
	i.AddEvent(event)

	return nil
}

func (i *ItemMenu) UpdateDescription(description string) error {

	if err := ValidateItemDescription(description); err != nil {
		return err
	}

	i.description = description
	event := NewItemMenuDescriptionChanged(i.ItemID)
	i.AddEvent(event)

	return nil
}

func (i *ItemMenu) Name() string             { return i.name }
func (i *ItemMenu) Description() string      { return i.description }
func (i *ItemMenu) BasePrice() common.Money  { return i.basePrice }
func (i *ItemMenu) Status() enums.ItemStatus { return i.status }

func (i *ItemMenu) AddEvent(event common.DomainEvent) {
	i.events = append(i.events, event)
}

func (i *ItemMenu) PullEvent() []common.DomainEvent {
	events := i.events
	i.events = nil
	return events
}

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
		events:      []common.DomainEvent{},
	}
}
