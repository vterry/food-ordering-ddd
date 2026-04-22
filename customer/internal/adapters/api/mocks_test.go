package api

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/customer/internal/core/domain/cart"
	"github.com/vterry/food-project/customer/internal/core/domain/customer"
	"github.com/vterry/food-project/customer/internal/core/ports"
)

type MockCustomerUseCase struct {
	RegisterFunc func(ctx context.Context, cmd ports.RegisterCustomerCommand) (vo.ID, error)
	AddAddrFunc  func(ctx context.Context, customerID vo.ID, cmd ports.AddAddressCommand) error
	GetCustFunc  func(ctx context.Context, customerID vo.ID) (*customer.Customer, error)
}

func (m *MockCustomerUseCase) RegisterCustomer(ctx context.Context, cmd ports.RegisterCustomerCommand) (vo.ID, error) {
	return m.RegisterFunc(ctx, cmd)
}

func (m *MockCustomerUseCase) AddAddress(ctx context.Context, customerID vo.ID, cmd ports.AddAddressCommand) error {
	return m.AddAddrFunc(ctx, customerID, cmd)
}

func (m *MockCustomerUseCase) GetCustomer(ctx context.Context, customerID vo.ID) (*customer.Customer, error) {
	return m.GetCustFunc(ctx, customerID)
}

type MockCartUseCase struct {
	GetCartFunc   func(ctx context.Context, customerID vo.ID) (*cart.Cart, error)
	AddItemFunc   func(ctx context.Context, customerID vo.ID, cmd ports.AddItemToCartCommand) error
	UpdateQtyFunc func(ctx context.Context, customerID vo.ID, cmd ports.UpdateCartItemCommand) error
	RemoveItemFunc func(ctx context.Context, customerID vo.ID, productID vo.ID) error
	ClearFunc     func(ctx context.Context, customerID vo.ID) error
	CheckoutFunc  func(ctx context.Context, customerID vo.ID) error
}

func (m *MockCartUseCase) GetCart(ctx context.Context, customerID vo.ID) (*cart.Cart, error) {
	return m.GetCartFunc(ctx, customerID)
}

func (m *MockCartUseCase) AddItemToCart(ctx context.Context, customerID vo.ID, cmd ports.AddItemToCartCommand) error {
	return m.AddItemFunc(ctx, customerID, cmd)
}

func (m *MockCartUseCase) UpdateItemQuantity(ctx context.Context, customerID vo.ID, cmd ports.UpdateCartItemCommand) error {
	return m.UpdateQtyFunc(ctx, customerID, cmd)
}

func (m *MockCartUseCase) RemoveItemFromCart(ctx context.Context, customerID vo.ID, productID vo.ID) error {
	return m.RemoveItemFunc(ctx, customerID, productID)
}

func (m *MockCartUseCase) ClearCart(ctx context.Context, customerID vo.ID) error {
	return m.ClearFunc(ctx, customerID)
}

func (m *MockCartUseCase) Checkout(ctx context.Context, customerID vo.ID) error {
	return m.CheckoutFunc(ctx, customerID)
}
