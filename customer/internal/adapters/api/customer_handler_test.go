package api

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	apigen "github.com/vterry/food-project/customer/internal/adapters/api/generated"
	"github.com/vterry/food-project/customer/internal/core/domain/cart"
	"github.com/vterry/food-project/customer/internal/core/domain/customer"
	"github.com/vterry/food-project/customer/internal/core/ports"
	"testing"
)

func TestCustomerHandler_RegisterCustomer(t *testing.T) {
	mockCust := &MockCustomerUseCase{
		RegisterFunc: func(ctx context.Context, cmd ports.RegisterCustomerCommand) (vo.ID, error) {
			return vo.NewID("123"), nil
		},
	}
	h := NewCustomerHandler(mockCust, nil)

	req := apigen.RegisterCustomerRequestObject{
		Body: &apigen.RegisterCustomerJSONRequestBody{
			Name:  "John",
			Email: "john@ex.com",
			Phone: "123",
		},
	}

	resp, err := h.RegisterCustomer(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r201, ok := resp.(apigen.RegisterCustomer201JSONResponse)
	if !ok {
		t.Fatalf("expected 201 response, got %T", resp)
	}
	if *r201.Id != "123" {
		t.Errorf("expected id 123, got %s", *r201.Id)
	}
}

func TestCustomerHandler_GetCustomer(t *testing.T) {
	mockCust := &MockCustomerUseCase{
		GetCustFunc: func(ctx context.Context, id vo.ID) (*customer.Customer, error) {
			name, _ := customer.NewName("John")
			email, _ := customer.NewEmail("john@ex.com")
			phone, _ := customer.NewPhone("123")
			return customer.NewCustomer(id, name, email, phone), nil
		},
	}
	h := NewCustomerHandler(mockCust, nil)

	req := apigen.GetCustomerRequestObject{Id: "123"}
	resp, err := h.GetCustomer(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r200, ok := resp.(apigen.GetCustomer200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if *r200.Name != "John" {
		t.Errorf("expected name John, got %s", *r200.Name)
	}
}

func TestCustomerHandler_AddItemToCart(t *testing.T) {
	mockCart := &MockCartUseCase{
		AddItemFunc: func(ctx context.Context, customerID vo.ID, cmd ports.AddItemToCartCommand) error {
			return nil
		},
	}
	h := NewCustomerHandler(nil, mockCart)

	req := apigen.AddItemToCartRequestObject{
		Id: "cust1",
		Body: &apigen.AddItemToCartJSONRequestBody{
			RestaurantId: "rest1",
			ProductId:    "prod1",
			Name:         "Pizza",
			Price:        10.0,
			Quantity:     1,
		},
	}

	resp, err := h.AddItemToCart(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := resp.(apigen.AddItemToCart200Response)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
}

func TestCustomerHandler_Checkout(t *testing.T) {
	mockCart := &MockCartUseCase{
		CheckoutFunc: func(ctx context.Context, customerID vo.ID) error {
			return nil
		},
	}
	h := NewCustomerHandler(nil, mockCart)

	req := apigen.CheckoutRequestObject{Id: "cust1"}
	resp, err := h.Checkout(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := resp.(apigen.Checkout202Response)
	if !ok {
		t.Fatalf("expected 202 response, got %T", resp)
	}
}

func TestCustomerHandler_AddAddress(t *testing.T) {
	mockCust := &MockCustomerUseCase{
		AddAddrFunc: func(ctx context.Context, id vo.ID, cmd ports.AddAddressCommand) error {
			return nil
		},
	}
	h := NewCustomerHandler(mockCust, nil)

	isDefault := true
	req := apigen.AddAddressRequestObject{
		Id: "cust1",
		Body: &apigen.AddAddressJSONRequestBody{
			Street:    "Main St",
			City:      "NY",
			ZipCode:   "10001",
			IsDefault: &isDefault,
		},
	}

	resp, err := h.AddAddress(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := resp.(apigen.AddAddress204Response)
	if !ok {
		t.Fatalf("expected 204 response, got %T", resp)
	}
}

func TestCustomerHandler_GetCart(t *testing.T) {
	mockCart := &MockCartUseCase{
		GetCartFunc: func(ctx context.Context, id vo.ID) (*cart.Cart, error) {
			return cart.NewCart(vo.NewID("cart1"), id), nil
		},
	}
	h := NewCustomerHandler(nil, mockCart)

	req := apigen.GetCartRequestObject{Id: "cust1"}
	resp, err := h.GetCart(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r200, ok := resp.(apigen.GetCart200JSONResponse)
	if !ok {
		t.Fatalf("expected 200 response, got %T", resp)
	}
	if *r200.CustomerId != "cust1" {
		t.Errorf("expected customer id cust1, got %s", *r200.CustomerId)
	}
}

func TestCustomerHandler_UpdateItemQuantity(t *testing.T) {
	mockCart := &MockCartUseCase{
		UpdateQtyFunc: func(ctx context.Context, id vo.ID, cmd ports.UpdateCartItemCommand) error {
			return nil
		},
	}
	h := NewCustomerHandler(nil, mockCart)

	req := apigen.UpdateItemQuantityRequestObject{
		Id:        "cust1",
		ProductId: "prod1",
		Body:      &apigen.UpdateItemQuantityJSONRequestBody{Quantity: 5},
	}

	resp, err := h.UpdateItemQuantity(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := resp.(apigen.UpdateItemQuantity204Response)
	if !ok {
		t.Fatalf("expected 204 response, got %T", resp)
	}
}

func TestCustomerHandler_RemoveItemFromCart(t *testing.T) {
	mockCart := &MockCartUseCase{
		RemoveItemFunc: func(ctx context.Context, customerID vo.ID, productID vo.ID) error {
			return nil
		},
	}
	h := NewCustomerHandler(nil, mockCart)

	req := apigen.RemoveItemFromCartRequestObject{Id: "cust1", ProductId: "prod1"}
	resp, err := h.RemoveItemFromCart(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := resp.(apigen.RemoveItemFromCart204Response)
	if !ok {
		t.Fatalf("expected 204 response, got %T", resp)
	}
}

func TestCustomerHandler_ClearCart(t *testing.T) {
	mockCart := &MockCartUseCase{
		ClearFunc: func(ctx context.Context, id vo.ID) error {
			return nil
		},
	}
	h := NewCustomerHandler(nil, mockCart)

	req := apigen.ClearCartRequestObject{Id: "cust1"}
	resp, err := h.ClearCart(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := resp.(apigen.ClearCart204Response)
	if !ok {
		t.Fatalf("expected 204 response, got %T", resp)
	}
}

