package domain

import (
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

type CustomerRegisteredEvent struct {
	base.BaseDomainEvent
	CustomerID vo.ID
	Name       string
	Email      string
}

func NewCustomerRegisteredEvent(customerID vo.ID, name, email string) CustomerRegisteredEvent {
	return CustomerRegisteredEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("event-" + customerID.String())),
		CustomerID:      customerID,
		Name:            name,
		Email:           email,
	}
}

func (e CustomerRegisteredEvent) EventType() string {
	return "CustomerRegistered"
}

type ItemAddedToCartEvent struct {
	base.BaseDomainEvent
	CustomerID   vo.ID
	RestaurantID vo.ID
	ProductID    vo.ID
	Quantity     int
}

func NewItemAddedToCartEvent(customerID, restaurantID, productID vo.ID, quantity int) ItemAddedToCartEvent {
	id := vo.NewID("event-cart-" + customerID.String())
	return ItemAddedToCartEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(id),
		CustomerID:      customerID,
		RestaurantID:    restaurantID,
		ProductID:       productID,
		Quantity:        quantity,
	}
}

func (e ItemAddedToCartEvent) EventType() string {
	return "ItemAddedToCart"
}
