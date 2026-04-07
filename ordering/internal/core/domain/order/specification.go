package order

import (
	"errors"

	common "github.com/vterry/food-ordering/common/pkg"
	"github.com/vterry/food-ordering/ordering/internal/core/domain/valueobjects"
)

var (
	ErrCustomerIdIsEmpty   = errors.New("customer id cannot be empty")
	ErrRestaurantIdIsEmpty = errors.New("restaurant id cannot be empty")
	ErrInvalidItemsSize    = errors.New("order must have at least one item")
)

type OrderParams struct {
	customerID      valueobjects.CustomerID
	restaurantID    valueobjects.RestaurantID
	deliveryAddress valueobjects.DeliveryAddress
	items           []valueobjects.OrderItem
}

func ValidateNewOrder(
	customerID valueobjects.CustomerID,
	restaurantID valueobjects.RestaurantID,
	deliveryAddress valueobjects.DeliveryAddress,
	items []valueobjects.OrderItem,
) error {
	params := OrderParams{
		customerID:      customerID,
		restaurantID:    restaurantID,
		deliveryAddress: deliveryAddress,
		items:           items,
	}
	spec := NewOrderSpecification()
	return spec(&params)
}

func NewOrderSpecification() common.Specification[OrderParams] {
	return common.And(
		ValidateCustomerIDSpec(),
		ValidateRestaurantIDSpec(),
		ValidateDeliveryAddressSpec(),
		ValidateItemsSpec(),
	)
}

func ValidateCustomerIDSpec() common.Specification[OrderParams] {
	return func(o *OrderParams) error {
		if o.customerID.IsZero() {
			return ErrCustomerIdIsEmpty
		}
		return nil
	}
}

func ValidateRestaurantIDSpec() common.Specification[OrderParams] {
	return func(o *OrderParams) error {
		if o.restaurantID.IsZero() {
			return ErrRestaurantIdIsEmpty
		}
		return nil
	}
}

func ValidateDeliveryAddressSpec() common.Specification[OrderParams] {
	return func(o *OrderParams) error {
		return o.deliveryAddress.Validate()
	}
}

func ValidateItemsSpec() common.Specification[OrderParams] {
	return func(o *OrderParams) error {
		if len(o.items) <= 0 {
			return ErrInvalidItemsSize
		}
		for _, item := range o.items {
			if err := item.Validate(); err != nil {
				return err
			}
		}
		return nil
	}
}
