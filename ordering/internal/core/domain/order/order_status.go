package order

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidStateTransition = errors.New("invalid state transition")
)

type OrderStatus string

const (
	StatusCreated                        OrderStatus = "CREATED"
	StatusAuthorizingPayment             OrderStatus = "AUTHORIZING_PAYMENT"
	StatusAwaitingRestaurantConfirmation OrderStatus = "AWAITING_RESTAURANT_CONFIRMATION"
	StatusCapturingPayment               OrderStatus = "CAPTURING_PAYMENT"
	StatusSchedulingDelivery             OrderStatus = "SCHEDULING_DELIVERY"
	StatusPreparing                      OrderStatus = "PREPARING"
	StatusReady                          OrderStatus = "READY"
	StatusOutForDelivery                 OrderStatus = "OUT_FOR_DELIVERY"
	StatusDelivered                      OrderStatus = "DELIVERED"
	StatusRejected                       OrderStatus = "REJECTED"
	StatusRestaurantRejected             OrderStatus = "RESTAURANT_REJECTED"
	StatusCaptureFailed                  OrderStatus = "CAPTURE_FAILED"
	StatusDeliveryRefused                OrderStatus = "DELIVERY_REFUSED"
	StatusCancelling                     OrderStatus = "CANCELLING"
	StatusCancelled                      OrderStatus = "CANCELLED"
)

func (s OrderStatus) String() string {
	return string(s)
}

// CanTransitionTo checks if the transition to the target state is valid.
func (s OrderStatus) CanTransitionTo(target OrderStatus) error {
	valid := false

	switch s {
	case StatusCreated:
		valid = target == StatusAuthorizingPayment
	case StatusAuthorizingPayment:
		valid = target == StatusAwaitingRestaurantConfirmation || target == StatusRejected
	case StatusAwaitingRestaurantConfirmation:
		valid = target == StatusCapturingPayment || target == StatusRestaurantRejected || target == StatusCancelling || target == StatusCancelled
	case StatusCapturingPayment:
		valid = target == StatusSchedulingDelivery || target == StatusCaptureFailed
	case StatusSchedulingDelivery:
		valid = target == StatusPreparing
	case StatusPreparing:
		valid = target == StatusReady
	case StatusReady:
		valid = target == StatusOutForDelivery
	case StatusOutForDelivery:
		valid = target == StatusDelivered || target == StatusDeliveryRefused
	case StatusRestaurantRejected, StatusCaptureFailed, StatusCancelling, StatusDeliveryRefused:
		valid = target == StatusCancelled
	case StatusRejected, StatusCancelled, StatusDelivered:
		valid = false // Terminal states
	}

	if !valid {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidStateTransition, s, target)
	}
	return nil
}
