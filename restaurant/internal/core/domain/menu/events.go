package menu

import (
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

type MenuActivatedEvent struct {
	base.BaseDomainEvent
	MenuID       vo.ID
	RestaurantID vo.ID
}

func NewMenuActivatedEvent(menuID, restaurantID vo.ID) MenuActivatedEvent {
	return MenuActivatedEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-menu-act-" + menuID.String())),
		MenuID:          menuID,
		RestaurantID:    restaurantID,
	}
}

func (e MenuActivatedEvent) EventType() string {
	return "restaurant.menu.activated"
}

type ItemAvailabilityChangedEvent struct {
	base.BaseDomainEvent
	MenuID      vo.ID
	ProductID   vo.ID
	IsAvailable bool
}

func NewItemAvailabilityChangedEvent(menuID, productID vo.ID, available bool) ItemAvailabilityChangedEvent {
	return ItemAvailabilityChangedEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-item-av-" + productID.String())),
		MenuID:          menuID,
		ProductID:       productID,
		IsAvailable:     available,
	}
}

func (e ItemAvailabilityChangedEvent) EventType() string {
	return "restaurant.menu.item_availability_changed"
}
