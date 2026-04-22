package ports

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/customer/internal/core/domain/cart"
	"github.com/vterry/food-project/customer/internal/core/domain/customer"
)

// Output Ports (Infrastructure Dependencies)
type CustomerRepository interface {
	Save(ctx context.Context, customer *customer.Customer) error
	FindByID(ctx context.Context, id vo.ID) (*customer.Customer, error)
	FindByEmail(ctx context.Context, email string) (*customer.Customer, error)
}

type CartRepository interface {
	Save(ctx context.Context, cart *cart.Cart) error
	FindByCustomerID(ctx context.Context, customerID vo.ID) (*cart.Cart, error)
	Delete(ctx context.Context, customerID vo.ID) error
}

type EventPublisher interface {
	Publish(ctx context.Context, events ...base.DomainEvent) error
}

type RestaurantMenuItem struct {
	ID           vo.ID
	RestaurantID vo.ID
	Name         string
	Price        vo.Money
	Available    bool
}

type RestaurantCatalogPort interface {
	GetMenuItem(ctx context.Context, restaurantID, itemID vo.ID) (*RestaurantMenuItem, error)
}
