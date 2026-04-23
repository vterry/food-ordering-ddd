package ticket

import (
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

type TicketConfirmedEvent struct {
	base.BaseDomainEvent
	TicketID vo.ID `json:"ticket_id"`
	OrderID  vo.ID `json:"order_id"`
}

func NewTicketConfirmedEvent(ticketID, orderID vo.ID) TicketConfirmedEvent {
	return TicketConfirmedEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-tkt-conf-" + ticketID.String())),
		TicketID:        ticketID,
		OrderID:         orderID,
	}
}

func (e TicketConfirmedEvent) EventType() string {
	return "restaurant.ticket.confirmed"
}

type TicketRejectedEvent struct {
	base.BaseDomainEvent
	TicketID vo.ID  `json:"ticket_id"`
	OrderID  vo.ID  `json:"order_id"`
	Reason   string `json:"reason"`
}

func NewTicketRejectedEvent(ticketID, orderID vo.ID, reason string) TicketRejectedEvent {
	return TicketRejectedEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-tkt-rej-" + ticketID.String())),
		TicketID:        ticketID,
		OrderID:         orderID,
		Reason:          reason,
	}
}

func (e TicketRejectedEvent) EventType() string {
	return "restaurant.ticket.rejected"
}

type TicketReadyEvent struct {
	base.BaseDomainEvent
	TicketID vo.ID `json:"ticket_id"`
	OrderID  vo.ID `json:"order_id"`
}

func NewTicketReadyEvent(ticketID, orderID vo.ID) TicketReadyEvent {
	return TicketReadyEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-tkt-ready-" + ticketID.String())),
		TicketID:        ticketID,
		OrderID:         orderID,
	}
}

func (e TicketReadyEvent) EventType() string {
	return "restaurant.ticket.ready"
}

type TicketCancelledEvent struct {
	base.BaseDomainEvent
	TicketID vo.ID `json:"ticket_id"`
	OrderID  vo.ID `json:"order_id"`
}

func NewTicketCancelledEvent(ticketID, orderID vo.ID) TicketCancelledEvent {
	return TicketCancelledEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("ev-tkt-can-" + ticketID.String())),
		TicketID:        ticketID,
		OrderID:         orderID,
	}
}

func (e TicketCancelledEvent) EventType() string {
	return "restaurant.ticket.cancelled"
}
