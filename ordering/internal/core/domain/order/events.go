package order

import (
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

type OrderCreatedEvent struct {
	base.BaseDomainEvent
	OrderID       vo.ID          `json:"order_id"`
	CustomerID    vo.ID          `json:"customer_id"`
	RestaurantID  vo.ID          `json:"restaurant_id"`
	TotalAmount   vo.Money       `json:"total_amount"`
	CorrelationID vo.ID          `json:"correlation_id"`
	Items         []OrderItemDTO `json:"items"`
}

type OrderItemDTO struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

func NewOrderCreatedEvent(o *Order) OrderCreatedEvent {
	items := make([]OrderItemDTO, len(o.items))
	for i, item := range o.items {
		items[i] = OrderItemDTO{
			ProductID: item.menuItemID.String(),
			Name:      item.name,
			Quantity:  item.quantity,
			Price:     item.unitPrice.Amount(),
		}
	}

	return OrderCreatedEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-ord-cre-" + o.ID().String())),
		OrderID:         o.ID(),
		CustomerID:      o.customerID,
		RestaurantID:    o.restaurantID,
		TotalAmount:     o.totalAmount,
		CorrelationID:   o.correlationID,
		Items:           items,
	}
}

func (e OrderCreatedEvent) EventType() string {
	return "ordering.order.created"
}

type OrderConfirmedEvent struct {
	base.BaseDomainEvent
	OrderID       vo.ID `json:"order_id"`
	CorrelationID vo.ID `json:"correlation_id"`
}

func NewOrderConfirmedEvent(orderID, correlationID vo.ID) OrderConfirmedEvent {
	return OrderConfirmedEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-ord-conf-" + orderID.String())),
		OrderID:         orderID,
		CorrelationID:   correlationID,
	}
}

func (e OrderConfirmedEvent) EventType() string {
	return "ordering.order.confirmed"
}

type OrderCancelledEvent struct {
	base.BaseDomainEvent
	OrderID       vo.ID  `json:"order_id"`
	Reason        string `json:"reason"`
	CancelledBy   string `json:"cancelled_by"`
	CorrelationID vo.ID  `json:"correlation_id"`
}

func NewOrderCancelledEvent(orderID vo.ID, reason, cancelledBy string, correlationID vo.ID) OrderCancelledEvent {
	return OrderCancelledEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-ord-can-" + orderID.String())),
		OrderID:         orderID,
		Reason:          reason,
		CancelledBy:     cancelledBy,
		CorrelationID:   correlationID,
	}
}

func (e OrderCancelledEvent) EventType() string {
	return "ordering.order.cancelled"
}

type OrderRejectedEvent struct {
	base.BaseDomainEvent
	OrderID       vo.ID  `json:"order_id"`
	Reason        string `json:"reason"`
	CorrelationID vo.ID  `json:"correlation_id"`
}

func NewOrderRejectedEvent(orderID vo.ID, reason string, correlationID vo.ID) OrderRejectedEvent {
	return OrderRejectedEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-ord-rej-" + orderID.String())),
		OrderID:         orderID,
		Reason:          reason,
		CorrelationID:   correlationID,
	}
}

func (e OrderRejectedEvent) EventType() string {
	return "ordering.order.rejected"
}

type OrderDeliveredEvent struct {
	base.BaseDomainEvent
	OrderID       vo.ID `json:"order_id"`
	CorrelationID vo.ID `json:"correlation_id"`
}

func NewOrderDeliveredEvent(orderID, correlationID vo.ID) OrderDeliveredEvent {
	return OrderDeliveredEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-ord-del-" + orderID.String())),
		OrderID:         orderID,
		CorrelationID:   correlationID,
	}
}

func (e OrderDeliveredEvent) EventType() string {
	return "ordering.order.delivered"
}

// Internal Saga Events (Published by Order Aggregate for SEC to react)

type OrderPaymentAuthorizedEvent struct {
	base.BaseDomainEvent
	OrderID       vo.ID `json:"order_id"`
	CorrelationID vo.ID `json:"correlation_id"`
}

func NewOrderPaymentAuthorizedEvent(orderID, correlationID vo.ID) OrderPaymentAuthorizedEvent {
	return OrderPaymentAuthorizedEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-ord-pay-auth-" + orderID.String())),
		OrderID:         orderID,
		CorrelationID:   correlationID,
	}
}

func (e OrderPaymentAuthorizedEvent) EventType() string {
	return "ordering.order.payment_authorized"
}

type OrderRestaurantConfirmedEvent struct {
	base.BaseDomainEvent
	OrderID       vo.ID `json:"order_id"`
	CorrelationID vo.ID `json:"correlation_id"`
}

func NewOrderRestaurantConfirmedEvent(orderID, correlationID vo.ID) OrderRestaurantConfirmedEvent {
	return OrderRestaurantConfirmedEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-ord-rest-conf-" + orderID.String())),
		OrderID:         orderID,
		CorrelationID:   correlationID,
	}
}

func (e OrderRestaurantConfirmedEvent) EventType() string {
	return "ordering.order.restaurant_confirmed"
}

type OrderRestaurantRejectedEvent struct {
	base.BaseDomainEvent
	OrderID       vo.ID  `json:"order_id"`
	Reason        string `json:"reason"`
	CorrelationID vo.ID  `json:"correlation_id"`
}

func NewOrderRestaurantRejectedEvent(orderID vo.ID, reason string, correlationID vo.ID) OrderRestaurantRejectedEvent {
	return OrderRestaurantRejectedEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-ord-rest-rej-" + orderID.String())),
		OrderID:         orderID,
		Reason:          reason,
		CorrelationID:   correlationID,
	}
}

func (e OrderRestaurantRejectedEvent) EventType() string {
	return "ordering.order.restaurant_rejected"
}

type OrderPaymentCaptureFailedEvent struct {
	base.BaseDomainEvent
	OrderID       vo.ID  `json:"order_id"`
	Reason        string `json:"reason"`
	CorrelationID vo.ID  `json:"correlation_id"`
}

func NewOrderPaymentCaptureFailedEvent(orderID vo.ID, reason string, correlationID vo.ID) OrderPaymentCaptureFailedEvent {
	return OrderPaymentCaptureFailedEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-ord-pay-cap-fail-" + orderID.String())),
		OrderID:         orderID,
		Reason:          reason,
		CorrelationID:   correlationID,
	}
}

func (e OrderPaymentCaptureFailedEvent) EventType() string {
	return "ordering.order.payment_capture_failed"
}

type OrderDeliveryRefusedEvent struct {
	base.BaseDomainEvent
	OrderID       vo.ID  `json:"order_id"`
	Reason        string `json:"reason"`
	CorrelationID vo.ID  `json:"correlation_id"`
}

func NewOrderDeliveryRefusedEvent(orderID vo.ID, reason string, correlationID vo.ID) OrderDeliveryRefusedEvent {
	return OrderDeliveryRefusedEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-ord-del-ref-" + orderID.String())),
		OrderID:         orderID,
		Reason:          reason,
		CorrelationID:   correlationID,
	}
}

func (e OrderDeliveryRefusedEvent) EventType() string {
	return "ordering.order.delivery_refused"
}
