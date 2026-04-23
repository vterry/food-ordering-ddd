package delivery

import (
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

type DeliveryScheduled struct {
	base.BaseDomainEvent
	DeliveryID   vo.ID   `json:"delivery_id"`
	OrderID      vo.ID   `json:"order_id"`
	RestaurantID vo.ID   `json:"restaurant_id"`
	CustomerID   vo.ID   `json:"customer_id"`
	Address      Address `json:"address"`
}

func (e DeliveryScheduled) EventType() string {
	return "delivery.scheduled"
}

type DeliveryPickedUp struct {
	base.BaseDomainEvent
	DeliveryID vo.ID       `json:"delivery_id"`
	Courier    CourierInfo `json:"courier"`
}

func (e DeliveryPickedUp) EventType() string {
	return "delivery.picked_up"
}

type DeliveryCompleted struct {
	base.BaseDomainEvent
	DeliveryID vo.ID `json:"delivery_id"`
}

func (e DeliveryCompleted) EventType() string {
	return "delivery.completed"
}

type DeliveryRefused struct {
	base.BaseDomainEvent
	DeliveryID vo.ID  `json:"delivery_id"`
	Reason     string `json:"reason"`
}

func (e DeliveryRefused) EventType() string {
	return "delivery.refused"
}

type DeliveryCancelled struct {
	base.BaseDomainEvent
	DeliveryID vo.ID  `json:"delivery_id"`
	Reason     string `json:"reason"`
}

func (e DeliveryCancelled) EventType() string {
	return "delivery.cancelled"
}
