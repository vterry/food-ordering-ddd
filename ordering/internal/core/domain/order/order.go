package order

import (
	"time"

	common "github.com/vterry/food-ordering/common/pkg"
	"github.com/vterry/food-ordering/ordering/internal/core/domain/enums"
	"github.com/vterry/food-ordering/ordering/internal/core/domain/valueobjects"
)

type Order struct {
	valueobjects.OrderID
	customerId      valueobjects.CustomerID
	restaurantID    valueobjects.RestaurantID
	deliveryAddress valueobjects.DeliveryAddress
	items           []valueobjects.OrderItem
	totalAmount     common.Money
	status          enums.OrderStatus
	paymentId       valueobjects.PaymentID
	deliveryId      valueobjects.DeliveryID
	createdAt       time.Time
	updatedAt       time.Time
	events          []common.DomainEvent
}

func NewOrder(
	customerID valueobjects.CustomerID,
	restaurantID valueobjects.RestaurantID,
	deliveryAddress valueobjects.DeliveryAddress,
	items []valueobjects.OrderItem,
) (*Order, error) {

	if err := ValidateNewOrder(customerID, restaurantID, deliveryAddress, items); err != nil {
		return nil, err
	}

	total := common.NewMoneyFromCents(0)
	for _, item := range items {
		total = total.Add(item.TotalPrice())
	}

	o := &Order{
		OrderID:         valueobjects.NewOrderID(),
		customerId:      customerID,
		restaurantID:    restaurantID,
		deliveryAddress: deliveryAddress,
		items:           items,
		totalAmount:     total,
		status:          enums.Pending,
		createdAt:       time.Now(),
	}
	// o.AddEvent(NewOrderPlaced(*o))
	return o, nil
}

// Permitir editar o endereço apenas com o STATUS PENDING
func (o *Order) ChangeDeliveryAddress(newDeliveryAddress valueobjects.DeliveryAddress) error {
	// validar se o status está peding, se sim permitir que o endereço seja atualizado, caso contrário retorna erro
	return nil
}

func (o *Order) MarkAsPaid(paymentId valueobjects.PaymentID) error {
	// O paymentID só pode ser adicionado no mesmo momento que o Statatus é atualizado para PAID
	return nil
}

func (o *Order) ReadyToDelivery(deliveryId valueobjects.DeliveryID) error {
	// O deliveryID só pode ser adicionar no mesmo momento que o Status é atualizado para IN_DELIVERY
	return nil
}

func (o *Order) MarkAsDelivered() error {
	return nil
}

func (o *Order) Confirm() error {
	return nil
}

// Think about keep reasons follow a pattern - like order status
func (o *Order) Cancel(reason string) error {
	return nil
}

func (o *Order) Fail(reason string) error {
	return nil
}

func Restore(
	orderId valueobjects.OrderID,
	customerId valueobjects.CustomerID,
	restaurantID valueobjects.RestaurantID,
	deliveryAddress valueobjects.DeliveryAddress,
	items []valueobjects.OrderItem,
	totalAmount common.Money,
	status enums.OrderStatus,
	paymentId valueobjects.PaymentID,
	deliveryId valueobjects.DeliveryID,
	createdAt time.Time,
	updatedAt time.Time,
) *Order {
	return &Order{
		OrderID:         orderId,
		customerId:      customerId,
		restaurantID:    restaurantID,
		deliveryAddress: deliveryAddress,
		items:           items,
		totalAmount:     totalAmount,
		status:          status,
		paymentId:       paymentId,
		deliveryId:      deliveryId,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
		events:          []common.DomainEvent{},
	}
}

func (o *Order) AddEvent(events ...common.DomainEvent) {
	o.events = append(o.events, events...)
}

func (o *Order) Events() []common.DomainEvent {
	return o.events
}

func (o *Order) ClearEvents() {
	o.events = nil
}

func (o *Order) PullEvent() []common.DomainEvent {
	events := o.events
	o.events = nil
	return events
}
