package services

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/customer/internal/core/domain/cart"
	"github.com/vterry/food-project/customer/internal/core/domain/customer"
	"github.com/vterry/food-project/customer/internal/core/ports"
)

type MockCustomerRepo struct {
	customers map[string]*customer.Customer
	SaveErr   error
	FindErr   error
	SaveCalled int
	FindCalled int
}

func NewMockCustomerRepo() *MockCustomerRepo {
	return &MockCustomerRepo{
		customers: make(map[string]*customer.Customer),
	}
}

func (m *MockCustomerRepo) Save(ctx context.Context, c *customer.Customer) error {
	m.SaveCalled++
	if m.SaveErr != nil {
		return m.SaveErr
	}
	m.customers[c.ID().String()] = c
	return nil
}

func (m *MockCustomerRepo) FindByID(ctx context.Context, id vo.ID) (*customer.Customer, error) {
	m.FindCalled++
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	return m.customers[id.String()], nil
}

func (m *MockCustomerRepo) FindByEmail(ctx context.Context, email string) (*customer.Customer, error) {
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	for _, c := range m.customers {
		if c.Email().String() == email {
			return c, nil
		}
	}
	return nil, nil
}

type MockCartRepo struct {
	carts      map[string]*cart.Cart
	SaveErr    error
	FindErr    error
	DeleteErr  error
	SaveCalled int
	FindCalled int
}

func NewMockCartRepo() *MockCartRepo {
	return &MockCartRepo{
		carts: make(map[string]*cart.Cart),
	}
}

func (m *MockCartRepo) Save(ctx context.Context, c *cart.Cart) error {
	m.SaveCalled++
	if m.SaveErr != nil {
		return m.SaveErr
	}
	m.carts[c.CustomerID().String()] = c
	return nil
}

func (m *MockCartRepo) FindByCustomerID(ctx context.Context, customerID vo.ID) (*cart.Cart, error) {
	m.FindCalled++
	if m.FindErr != nil {
		return nil, m.FindErr
	}
	return m.carts[customerID.String()], nil
}

func (m *MockCartRepo) Delete(ctx context.Context, customerID vo.ID) error {
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	delete(m.carts, customerID.String())
	return nil
}

type MockPublisher struct {
	Events     []base.DomainEvent
	PublishErr error
}

func (m *MockPublisher) Publish(ctx context.Context, events ...base.DomainEvent) error {
	if m.PublishErr != nil {
		return m.PublishErr
	}
	m.Events = append(m.Events, events...)
	return nil
}

type MockRestaurantCatalogPort struct {
	Items  map[string]*ports.RestaurantMenuItem
	GetErr error
}

func NewMockRestaurantCatalogPort() *MockRestaurantCatalogPort {
	return &MockRestaurantCatalogPort{
		Items: make(map[string]*ports.RestaurantMenuItem),
	}
}

func (m *MockRestaurantCatalogPort) GetMenuItem(ctx context.Context, restaurantID, itemID vo.ID) (*ports.RestaurantMenuItem, error) {
	if m.GetErr != nil {
		return nil, m.GetErr
	}
	return m.Items[itemID.String()], nil
}

func (m *MockRestaurantCatalogPort) AddItem(item *ports.RestaurantMenuItem) {
	m.Items[item.ID.String()] = item
}
