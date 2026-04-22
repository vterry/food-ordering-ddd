package ports

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/restaurant/internal/core/domain/restaurant"
	"github.com/vterry/food-project/restaurant/internal/core/domain/ticket"
)

// Commands for Restaurant
type CreateRestaurantCommand struct {
	Name    string
	Address restaurant.Address
	Hours   []restaurant.OperatingPeriod
}

type AddMenuItemCommand struct {
	MenuID      vo.ID
	Name        string
	Description string
	Price       vo.Money
	Category    string
}

// Commands for Ticket
type CreateTicketCommand struct {
	OrderID      vo.ID
	RestaurantID vo.ID
	Items        []ticket.TicketItem
}

// Input Ports (Use Cases)
type RestaurantUseCase interface {
	CreateRestaurant(ctx context.Context, cmd CreateRestaurantCommand) (vo.ID, error)
	GetRestaurant(ctx context.Context, id vo.ID) (*restaurant.Restaurant, error)
	CreateMenu(ctx context.Context, restaurantID vo.ID, name string) (vo.ID, error)
	ActivateMenu(ctx context.Context, menuID vo.ID) error
	AddItemToMenu(ctx context.Context, cmd AddMenuItemCommand) (vo.ID, error)
	UpdateItemAvailability(ctx context.Context, menuID, productID vo.ID, available bool) error
}

type TicketUseCase interface {
	CreateTicket(ctx context.Context, cmd CreateTicketCommand) (vo.ID, error)
	ConfirmTicket(ctx context.Context, ticketID vo.ID) error
	StartPreparingTicket(ctx context.Context, ticketID vo.ID) error
	RejectTicket(ctx context.Context, ticketID vo.ID, reason string) error
	MarkTicketAsReady(ctx context.Context, ticketID vo.ID) error
	GetTicket(ctx context.Context, ticketID vo.ID) (*ticket.Ticket, error)
}
