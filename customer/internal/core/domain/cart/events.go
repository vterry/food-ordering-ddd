package cart

import (
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

type ItemAddedToCartEvent struct {
	base.BaseDomainEvent
	CustomerID   vo.ID
	RestaurantID vo.ID
	ProductID    vo.ID
	Quantity     int
}

func NewItemAddedToCartEvent(customerID, restaurantID, productID vo.ID, quantity int) ItemAddedToCartEvent {
	return ItemAddedToCartEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("event-" + customerID.String())),
		CustomerID:      customerID,
		RestaurantID:    restaurantID,
		ProductID:       productID,
		Quantity:        quantity,
	}
}

type CheckoutRequestedEvent struct {
	base.BaseDomainEvent
	CustomerID   vo.ID
	RestaurantID vo.ID
	Items        []CartItem
}

func (e CheckoutRequestedEvent) EventType() string {
	return "cart.checkout_requested"
}

func NewCheckoutRequestedEvent(customerID, restaurantID vo.ID, items []CartItem) CheckoutRequestedEvent {
	return CheckoutRequestedEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("event-checkout-" + customerID.String())),
		CustomerID:      customerID,
		RestaurantID:    restaurantID,
		Items:           items,
	}
}

func (e ItemAddedToCartEvent) EventType() string {
	return "ItemAddedToCart"
}
