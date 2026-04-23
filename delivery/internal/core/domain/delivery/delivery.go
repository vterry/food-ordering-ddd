package delivery

import (
	"time"
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

type Delivery struct {
	base.BaseAggregateRoot
	orderID       vo.ID
	restaurantID  vo.ID
	customerID    vo.ID
	address       Address
	status        Status
	courier       *CourierInfo
	correlationID vo.ID
	createdAt     time.Time
	updatedAt     time.Time
}

func NewDelivery(id vo.ID, orderID vo.ID, restaurantID vo.ID, customerID vo.ID, address Address, correlationID vo.ID) *Delivery {
	d := &Delivery{
		orderID:       orderID,
		restaurantID:  restaurantID,
		customerID:    customerID,
		address:       address,
		status:        StatusScheduled,
		correlationID: correlationID,
		createdAt:     time.Now(),
		updatedAt:     time.Now(),
	}
	d.SetID(id)

	d.AddEvent(DeliveryScheduled{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("")), // ID generation usually happens at infrastructure or common level
		DeliveryID:      id,
		OrderID:         orderID,
		RestaurantID:    restaurantID,
		CustomerID:      customerID,
		Address:         address,
	})

	return d
}

func (d *Delivery) OrderID() vo.ID {
	return d.orderID
}

func (d *Delivery) RestaurantID() vo.ID {
	return d.restaurantID
}

func (d *Delivery) CustomerID() vo.ID {
	return d.customerID
}

func (d *Delivery) Address() Address {
	return d.address
}

func (d *Delivery) Status() Status {
	return d.status
}

func (d *Delivery) Courier() *CourierInfo {
	return d.courier
}

func (d *Delivery) CorrelationID() vo.ID {
	return d.correlationID
}

func (d *Delivery) CreatedAt() time.Time {
	return d.createdAt
}

func (d *Delivery) UpdatedAt() time.Time {
	return d.updatedAt
}

func MapFromPersistence(
	id vo.ID,
	orderID vo.ID,
	restaurantID vo.ID,
	customerID vo.ID,
	address Address,
	status Status,
	courier *CourierInfo,
	correlationID vo.ID,
	createdAt time.Time,
	updatedAt time.Time,
) *Delivery {
	d := &Delivery{
		orderID:       orderID,
		restaurantID:  restaurantID,
		customerID:    customerID,
		address:       address,
		status:        status,
		courier:       courier,
		correlationID: correlationID,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
	d.SetID(id)
	return d
}

func (d *Delivery) PickUp(courier CourierInfo) error {
	if d.status != StatusScheduled {
		return ErrInvalidStatusTransition
	}

	d.status = StatusPickedUp
	d.courier = &courier

	d.AddEvent(DeliveryPickedUp{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("")),
		DeliveryID:      d.ID(),
		Courier:         courier,
	})

	return nil
}

func (d *Delivery) Complete() error {
	if d.status != StatusPickedUp {
		return ErrDeliveryNotPickedUp
	}

	d.status = StatusDelivered

	d.AddEvent(DeliveryCompleted{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("")),
		DeliveryID:      d.ID(),
	})

	return nil
}

func (d *Delivery) Refuse(reason string) error {
	if d.status != StatusPickedUp {
		return ErrDeliveryNotPickedUp
	}

	d.status = StatusRefused

	d.AddEvent(DeliveryRefused{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("")),
		DeliveryID:      d.ID(),
		Reason:          reason,
	})

	return nil
}

func (d *Delivery) Cancel(reason string) error {
	if d.status == StatusDelivered || d.status == StatusRefused {
		return ErrInvalidStatusTransition
	}

	d.status = StatusCancelled

	d.AddEvent(DeliveryCancelled{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("")),
		DeliveryID:      d.ID(),
		Reason:          reason,
	})

	return nil
}
