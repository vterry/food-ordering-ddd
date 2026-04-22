package services

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/restaurant/internal/core/domain/menu"
	"github.com/vterry/food-project/restaurant/internal/core/domain/restaurant"
	"github.com/vterry/food-project/restaurant/internal/core/domain/ticket"
)

type MockRestaurantRepo struct {
	items map[string]*restaurant.Restaurant
}

func NewMockRestaurantRepo() *MockRestaurantRepo {
	return &MockRestaurantRepo{items: make(map[string]*restaurant.Restaurant)}
}

func (m *MockRestaurantRepo) Save(ctx context.Context, r *restaurant.Restaurant) error {
	m.items[r.ID().String()] = r
	return nil
}

func (m *MockRestaurantRepo) FindByID(ctx context.Context, id vo.ID) (*restaurant.Restaurant, error) {
	return m.items[id.String()], nil
}

type MockMenuRepo struct {
	items map[string]*menu.Menu
}

func NewMockMenuRepo() *MockMenuRepo {
	return &MockMenuRepo{items: make(map[string]*menu.Menu)}
}

func (m *MockMenuRepo) Save(ctx context.Context, menu *menu.Menu) error {
	m.items[menu.ID().String()] = menu
	return nil
}

func (m *MockMenuRepo) FindByID(ctx context.Context, id vo.ID) (*menu.Menu, error) {
	return m.items[id.String()], nil
}

func (m *MockMenuRepo) FindActiveByRestaurantID(ctx context.Context, restaurantID vo.ID) (*menu.Menu, error) {
	for _, menu := range m.items {
		if menu.RestaurantID().Equals(restaurantID) && menu.IsActive() {
			return menu, nil
		}
	}
	return nil, nil
}

type MockTicketRepo struct {
	items map[string]*ticket.Ticket
}

func NewMockTicketRepo() *MockTicketRepo {
	return &MockTicketRepo{items: make(map[string]*ticket.Ticket)}
}

func (m *MockTicketRepo) Save(ctx context.Context, t *ticket.Ticket) error {
	m.items[t.ID().String()] = t
	return nil
}

func (m *MockTicketRepo) FindByID(ctx context.Context, id vo.ID) (*ticket.Ticket, error) {
	return m.items[id.String()], nil
}
