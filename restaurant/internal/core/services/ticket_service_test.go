package services

import (
	"context"
	"testing"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/restaurant/internal/core/domain/ticket"
	"github.com/vterry/food-project/restaurant/internal/core/ports"
)

func TestTicketService_Lifecycle(t *testing.T) {
	repo := NewMockTicketRepo()
	svc := NewTicketService(repo)

	orderID := vo.NewID("ord-1")
	restID := vo.NewID("rest-1")

	cmd := ports.CreateTicketCommand{
		OrderID:      orderID,
		RestaurantID: restID,
		Items: []ticket.TicketItem{
			{ProductID: vo.NewID("p1"), Name: "Pizza", Quantity: 1},
		},
	}

	id, err := svc.CreateTicket(context.Background(), cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Confirm
	err = svc.ConfirmTicket(context.Background(), id)
	if err != nil {
		t.Errorf("unexpected error confirming: %v", err)
	}

	tk, _ := repo.FindByID(context.Background(), id)
	if tk.Status() != ticket.StatusConfirmed {
		t.Errorf("expected status CONFIRMED, got %s", tk.Status())
	}

	// Preparing
	err = svc.StartPreparingTicket(context.Background(), id)
	if err != nil {
		t.Errorf("unexpected error starting preparing: %v", err)
	}
	if tk.Status() != ticket.StatusPreparing {
		t.Errorf("expected status PREPARING, got %s", tk.Status())
	}

	// Ready
	err = svc.MarkTicketAsReady(context.Background(), id)
	if err != nil {
		t.Errorf("unexpected error marking as ready: %v", err)
	}

	if tk.Status() != ticket.StatusReady {
		t.Errorf("expected status READY, got %s", tk.Status())
	}
}
