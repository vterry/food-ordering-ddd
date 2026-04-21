package ports

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"github.com/vterry/food-project/customer/internal/core/domain/cart"
	"github.com/vterry/food-project/customer/internal/core/domain/customer"
)

// Commands for Customer
type RegisterCustomerCommand struct {
	Name  string
	Email string
	Phone string
}

type AddAddressCommand struct {
	Street    string
	City      string
	ZipCode   string
	IsDefault bool
}

// Commands for Cart
type AddItemToCartCommand struct {
	RestaurantID vo.ID
	ProductID    vo.ID
	Name         string
	Price        vo.Money
	Quantity     int
	Observation  string
}

type UpdateCartItemCommand struct {
	ProductID vo.ID
	Quantity  int
}

// Input Ports (Use Cases)
type CustomerUseCase interface {
	RegisterCustomer(ctx context.Context, cmd RegisterCustomerCommand) (vo.ID, error)
	AddAddress(ctx context.Context, customerID vo.ID, cmd AddAddressCommand) error
	GetCustomer(ctx context.Context, customerID vo.ID) (*customer.Customer, error)
}

type CartUseCase interface {
	GetCart(ctx context.Context, customerID vo.ID) (*cart.Cart, error)
	AddItemToCart(ctx context.Context, customerID vo.ID, cmd AddItemToCartCommand) error
	UpdateItemQuantity(ctx context.Context, customerID vo.ID, cmd UpdateCartItemCommand) error
	RemoveItemFromCart(ctx context.Context, customerID vo.ID, productID vo.ID) error
	ClearCart(ctx context.Context, customerID vo.ID) error
	Checkout(ctx context.Context, customerID vo.ID) error
}
