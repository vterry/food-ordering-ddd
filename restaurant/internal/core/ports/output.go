package ports

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/restaurant/internal/core/domain/menu"
	"github.com/vterry/food-project/restaurant/internal/core/domain/restaurant"
	"github.com/vterry/food-project/restaurant/internal/core/domain/ticket"
)

type RestaurantRepository interface {
	Save(ctx context.Context, r *restaurant.Restaurant) error
	FindByID(ctx context.Context, id vo.ID) (*restaurant.Restaurant, error)
}

type MenuRepository interface {
	Save(ctx context.Context, m *menu.Menu) error
	FindByID(ctx context.Context, id vo.ID) (*menu.Menu, error)
	FindActiveByRestaurantID(ctx context.Context, restaurantID vo.ID) (*menu.Menu, error)
}

type TicketRepository interface {
	Save(ctx context.Context, t *ticket.Ticket) error
	FindByID(ctx context.Context, id vo.ID) (*ticket.Ticket, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, events ...base.DomainEvent) error
}
