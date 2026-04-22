package ticket

import (
	"errors"
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

type TicketStatus string

const (
	StatusPending   TicketStatus = "PENDING"
	StatusConfirmed TicketStatus = "CONFIRMED"
	StatusPreparing TicketStatus = "PREPARING"
	StatusReady     TicketStatus = "READY"
	StatusRejected  TicketStatus = "REJECTED"
	StatusCancelled TicketStatus = "CANCELLED"
)

var (
	ErrInvalidStateTransition = errors.New("invalid status transition")
)

type TicketItem struct {
	ProductID vo.ID
	Name      string
	Quantity  int
}

type Ticket struct {
	base.BaseAggregateRoot
	orderID      vo.ID
	restaurantID vo.ID
	status       TicketStatus
	items        []TicketItem
	rejectReason string
}

func NewTicket(id, orderID, restaurantID vo.ID, items []TicketItem) *Ticket {
	t := &Ticket{
		orderID:      orderID,
		restaurantID: restaurantID,
		status:       StatusPending,
		items:        items,
	}
	t.SetID(id)
	return t
}

func (t *Ticket) OrderID() vo.ID {
	return t.orderID
}

func (t *Ticket) RestaurantID() vo.ID {
	return t.restaurantID
}

func (t *Ticket) Status() TicketStatus {
	return t.status
}

func (t *Ticket) Items() []TicketItem {
	return t.items
}

// MapFromPersistence reconstructs a ticket from persistence data.
func MapFromPersistence(id, orderID, restaurantID vo.ID, status TicketStatus, rejectReason string, items []TicketItem) *Ticket {
	t := &Ticket{
		orderID:      orderID,
		restaurantID: restaurantID,
		status:       status,
		items:        items,
		rejectReason: rejectReason,
	}
	t.SetID(id)
	return t
}

func (t *Ticket) Confirm() error {
	if t.status != StatusPending {
		return ErrInvalidStateTransition
	}
	t.status = StatusConfirmed
	t.AddEvent(NewTicketConfirmedEvent(t.ID(), t.orderID))
	return nil
}

func (t *Ticket) Reject(reason string) error {
	if t.status != StatusPending {
		return ErrInvalidStateTransition
	}
	t.status = StatusRejected
	t.rejectReason = reason
	t.AddEvent(NewTicketRejectedEvent(t.ID(), t.orderID, reason))
	return nil
}

func (t *Ticket) MarkAsPreparing() error {
	if t.status != StatusConfirmed {
		return ErrInvalidStateTransition
	}
	t.status = StatusPreparing
	return nil
}

func (t *Ticket) MarkAsReady() error {
	if t.status != StatusPreparing {
		return ErrInvalidStateTransition
	}
	t.status = StatusReady
	t.AddEvent(NewTicketReadyEvent(t.ID(), t.orderID))
	return nil
}

func (t *Ticket) Cancel() {
	t.status = StatusCancelled
	t.AddEvent(NewTicketCancelledEvent(t.ID(), t.orderID))
}
