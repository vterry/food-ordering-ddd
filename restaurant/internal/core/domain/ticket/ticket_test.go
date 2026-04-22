package ticket

import (
	"testing"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

func TestNewTicket(t *testing.T) {
	id := vo.NewID("t-1")
	orderID := vo.NewID("o-1")
	restaurantID := vo.NewID("r-1")
	items := []TicketItem{
		{ProductID: vo.NewID("p-1"), Name: "Pizza", Quantity: 1},
	}

	ticket := NewTicket(id, orderID, restaurantID, items)

	if !ticket.ID().Equals(id) {
		t.Errorf("expected ID %v, got %v", id, ticket.ID())
	}
	if ticket.Status() != StatusPending {
		t.Errorf("expected status PENDING, got %s", ticket.Status())
	}
}

func TestTicket_StateTransitions(t *testing.T) {
	ticket := NewTicket(vo.NewID("t-1"), vo.NewID("o-1"), vo.NewID("r-1"), nil)

	// PENDING -> CONFIRMED
	err := ticket.Confirm()
	if err != nil {
		t.Errorf("unexpected error confirming ticket: %v", err)
	}
	if len(ticket.Events()) != 1 || ticket.Events()[0].EventType() != "restaurant.ticket.confirmed" {
		t.Errorf("expected confirmed event")
	}

	// CONFIRMED -> PREPARING
	ticket.ClearEvents()
	err = ticket.MarkAsPreparing()
	if err != nil {
		t.Errorf("unexpected error marking as preparing: %v", err)
	}
	if len(ticket.Events()) != 0 {
		t.Errorf("no event expected for preparing status")
	}

	// PREPARING -> READY
	err = ticket.MarkAsReady()
	if err != nil {
		t.Errorf("unexpected error marking as ready: %v", err)
	}
	if len(ticket.Events()) != 1 || ticket.Events()[0].EventType() != "restaurant.ticket.ready" {
		t.Errorf("expected ready event")
	}
}

func TestTicket_Reject(t *testing.T) {
	ticket := NewTicket(vo.NewID("t-1"), vo.NewID("o-1"), vo.NewID("r-1"), nil)
	err := ticket.Reject("out of stock")
	if err != nil {
		t.Errorf("unexpected error rejecting: %v", err)
	}
	if len(ticket.Events()) != 1 || ticket.Events()[0].EventType() != "restaurant.ticket.rejected" {
		t.Errorf("expected rejected event")
	}
}

func TestTicket_Cancel(t *testing.T) {
	ticket := NewTicket(vo.NewID("t-1"), vo.NewID("o-1"), vo.NewID("r-1"), nil)
	ticket.Cancel()
	if len(ticket.Events()) != 1 || ticket.Events()[0].EventType() != "restaurant.ticket.cancelled" {
		t.Errorf("expected cancelled event")
	}
}

func TestTicket_InvalidTransition(t *testing.T) {
	ticket := NewTicket(vo.NewID("t-1"), vo.NewID("o-1"), vo.NewID("r-1"), nil)
	ticket.Confirm()
	
	// Try to reject after confirmed (invalid)
	err := ticket.Reject("too late")
	if err == nil {
		t.Errorf("expected error rejecting confirmed ticket, got nil")
	}
}
