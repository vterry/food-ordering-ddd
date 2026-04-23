package order

import (
	"fmt"
	"time"

	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

// Order is the aggregate root for the Ordering bounded context.
type Order struct {
	base.BaseAggregateRoot
	customerID     vo.ID
	restaurantID   vo.ID
	items          []OrderItem
	status         OrderStatus
	totalAmount    vo.Money
	deliveryAddress string // Simplified for now
	correlationID  vo.ID
	version        int
	createdAt      time.Time
	updatedAt      time.Time
}

// NewOrder creates a new Order aggregate.
func NewOrder(id, customerID, restaurantID vo.ID, items []OrderItem, deliveryAddress string, correlationID vo.ID) (*Order, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("order must have at least one item")
	}

	total, err := calculateTotal(items)
	if err != nil {
		return nil, err
	}

	o := &Order{
		customerID:      customerID,
		restaurantID:    restaurantID,
		items:           items,
		status:          StatusCreated,
		totalAmount:     total,
		deliveryAddress: deliveryAddress,
		correlationID:   correlationID,
		version:         1,
		createdAt:       time.Now(),
		updatedAt:       time.Now(),
	}
	o.SetID(id)
	
	// Register OrderCreated event
	o.AddEvent(NewOrderCreatedEvent(o))
	
	return o, nil
}

func calculateTotal(items []OrderItem) (vo.Money, error) {
	if len(items) == 0 {
		return vo.Money{}, fmt.Errorf("cannot calculate total for empty items")
	}
	
	total := items[0].TotalPrice()
	for i := 1; i < len(items); i++ {
		var err error
		total, err = total.Add(items[i].TotalPrice())
		if err != nil {
			return vo.Money{}, err
		}
	}
	return total, nil
}

// Getters

func (o *Order) CustomerID() vo.ID { return o.customerID }
func (o *Order) RestaurantID() vo.ID { return o.restaurantID }
func (o *Order) Items() []OrderItem { return o.items }
func (o *Order) Status() OrderStatus { return o.status }
func (o *Order) TotalAmount() vo.Money { return o.totalAmount }
func (o *Order) DeliveryAddress() string { return o.deliveryAddress }
func (o *Order) CorrelationID() vo.ID { return o.correlationID }
func (o *Order) Version() int { return o.version }
func (o *Order) CreatedAt() time.Time { return o.createdAt }
func (o *Order) UpdatedAt() time.Time { return o.updatedAt }

// Transitions

func (o *Order) transitionTo(target OrderStatus) error {
	if err := o.status.CanTransitionTo(target); err != nil {
		return err
	}
	o.status = target
	o.version++
	o.updatedAt = time.Now()
	return nil
}

func (o *Order) MarkPaymentPending() error {
	return o.transitionTo(StatusAuthorizingPayment)
}

func (o *Order) MarkPaymentAuthorized() error {
	if err := o.transitionTo(StatusAwaitingRestaurantConfirmation); err != nil {
		return err
	}
	o.AddEvent(NewOrderPaymentAuthorizedEvent(o.ID(), o.correlationID))
	return nil
}

func (o *Order) Reject(reason string) error {
	if err := o.transitionTo(StatusRejected); err != nil {
		return err
	}
	o.AddEvent(NewOrderRejectedEvent(o.ID(), reason, o.correlationID))
	return nil
}

func (o *Order) MarkRestaurantConfirmed() error {
	if err := o.transitionTo(StatusCapturingPayment); err != nil {
		return err
	}
	o.AddEvent(NewOrderRestaurantConfirmedEvent(o.ID(), o.correlationID))
	return nil
}

func (o *Order) MarkRestaurantRejected(reason string) error {
	if err := o.transitionTo(StatusRestaurantRejected); err != nil {
		return err
	}
	o.AddEvent(NewOrderRestaurantRejectedEvent(o.ID(), reason, o.correlationID))
	return nil
}

func (o *Order) Cancel() error {
	if err := o.transitionTo(StatusCancelling); err != nil {
		return err
	}
	o.AddEvent(NewOrderCancelledEvent(o.ID(), "Cancelled by customer", "CUSTOMER", o.correlationID))
	return nil
}

func (o *Order) MarkPaymentCaptured() error {
	if err := o.transitionTo(StatusSchedulingDelivery); err != nil {
		return err
	}
	o.AddEvent(NewOrderConfirmedEvent(o.ID(), o.correlationID))
	return nil
}

func (o *Order) MarkCaptureFailed(reason string) error {
	if err := o.transitionTo(StatusCaptureFailed); err != nil {
		return err
	}
	o.AddEvent(NewOrderPaymentCaptureFailedEvent(o.ID(), reason, o.correlationID))
	return nil
}

func (o *Order) MarkDeliveryScheduled() error {
	return o.transitionTo(StatusPreparing)
}

func (o *Order) MarkReady() error {
	return o.transitionTo(StatusReady)
}

func (o *Order) MarkOutForDelivery() error {
	return o.transitionTo(StatusOutForDelivery)
}

func (o *Order) MarkDelivered() error {
	if err := o.transitionTo(StatusDelivered); err != nil {
		return err
	}
	o.AddEvent(NewOrderDeliveredEvent(o.ID(), o.correlationID))
	return nil
}

func (o *Order) MarkDeliveryRefused(reason string) error {
	if err := o.transitionTo(StatusDeliveryRefused); err != nil {
		return err
	}
	o.AddEvent(NewOrderDeliveryRefusedEvent(o.ID(), reason, o.correlationID))
	return nil
}

func (o *Order) MarkCancelled() error {
	return o.transitionTo(StatusCancelled)
}

// MapFromPersistence reconstructs the aggregate from persistence.
func MapFromPersistence(
	id vo.ID,
	customerID vo.ID,
	restaurantID vo.ID,
	items []OrderItem,
	status OrderStatus,
	totalAmount vo.Money,
	deliveryAddress string,
	correlationID vo.ID,
	version int,
	createdAt time.Time,
	updatedAt time.Time,
) *Order {
	o := &Order{
		customerID:      customerID,
		restaurantID:    restaurantID,
		items:           items,
		status:          status,
		totalAmount:     totalAmount,
		deliveryAddress: deliveryAddress,
		correlationID:   correlationID,
		version:         version,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
	o.SetID(id)
	return o
}
