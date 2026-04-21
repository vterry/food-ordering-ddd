package api

import (
	"context"
	"github.com/vterry/food-project/common/pkg/domain/vo"
	apigen "github.com/vterry/food-project/customer/internal/adapters/api/generated"
	"github.com/vterry/food-project/customer/internal/core/ports"
)

type CustomerHandler struct {
	customerSvc ports.CustomerUseCase
	cartSvc     ports.CartUseCase
}

func NewCustomerHandler(customerSvc ports.CustomerUseCase, cartSvc ports.CartUseCase) *CustomerHandler {
	return &CustomerHandler{
		customerSvc: customerSvc,
		cartSvc:     cartSvc,
	}
}

// Implementation of apigen.StrictServerInterface

func (h *CustomerHandler) RegisterCustomer(ctx context.Context, request apigen.RegisterCustomerRequestObject) (apigen.RegisterCustomerResponseObject, error) {
	id, err := h.customerSvc.RegisterCustomer(ctx, ports.RegisterCustomerCommand{
		Name:  request.Body.Name,
		Email: string(request.Body.Email),
		Phone: request.Body.Phone,
	})
	if err != nil {
		return nil, h.handleError(err)
	}
	sID := id.String()
	return apigen.RegisterCustomer201JSONResponse{Id: &sID}, nil
}

func (h *CustomerHandler) GetCustomer(ctx context.Context, request apigen.GetCustomerRequestObject) (apigen.GetCustomerResponseObject, error) {
	c, err := h.customerSvc.GetCustomer(ctx, vo.NewID(request.Id))
	if err != nil {
		return nil, h.handleError(err)
	}
	
	addrs := make([]apigen.Address, 0)
	for _, a := range c.Addresses() {
		addrs = append(addrs, apigen.Address{
			Id:        ptr(a.ID().String()),
			Street:    ptr(a.Street()),
			City:      ptr(a.City()),
			ZipCode:   ptr(a.ZipCode()),
			IsDefault: ptr(a.IsDefault()),
		})
	}

	return apigen.GetCustomer200JSONResponse{
		Id:        ptr(c.ID().String()),
		Name:      ptr(c.Name().String()),
		Email:     ptr(c.Email().String()),
		Phone:     ptr(c.Phone().String()),
		Addresses: &addrs,
	}, nil
}

func (h *CustomerHandler) AddAddress(ctx context.Context, request apigen.AddAddressRequestObject) (apigen.AddAddressResponseObject, error) {
	isDefault := false
	if request.Body.IsDefault != nil {
		isDefault = *request.Body.IsDefault
	}
	
	err := h.customerSvc.AddAddress(ctx, vo.NewID(request.Id), ports.AddAddressCommand{
		Street:    request.Body.Street,
		City:      request.Body.City,
		ZipCode:   request.Body.ZipCode,
		IsDefault: isDefault,
	})
	if err != nil {
		return nil, h.handleError(err)
	}
	return apigen.AddAddress204Response{}, nil
}

func (h *CustomerHandler) GetCart(ctx context.Context, request apigen.GetCartRequestObject) (apigen.GetCartResponseObject, error) {
	c, err := h.cartSvc.GetCart(ctx, vo.NewID(request.Id))
	if err != nil {
		return nil, h.handleError(err)
	}

	items := make([]apigen.CartItem, 0)
	for _, it := range c.Items() {
		fPrice := float32(it.Price().Amount())
		q := it.Quantity()
		items = append(items, apigen.CartItem{
			ProductId:   ptr(it.ProductID().String()),
			Name:        ptr(it.Name()),
			Price:       &fPrice,
			Quantity:    &q,
			Observation: ptr(it.Observation()),
		})
	}

	fTotal := float32(c.TotalValue().Amount())
	return apigen.GetCart200JSONResponse{
		CustomerId:   ptr(c.CustomerID().String()),
		RestaurantId: ptr(c.RestaurantID().String()),
		Items:        &items,
		Total:        &fTotal,
	}, nil
}

func (h *CustomerHandler) AddItemToCart(ctx context.Context, request apigen.AddItemToCartRequestObject) (apigen.AddItemToCartResponseObject, error) {
	currency := "BRL"
	if request.Body.Currency != nil {
		currency = *request.Body.Currency
	}
	
	price, _ := vo.NewMoney(float64(request.Body.Price), currency)
	
	obs := ""
	if request.Body.Observation != nil {
		obs = *request.Body.Observation
	}

	err := h.cartSvc.AddItemToCart(ctx, vo.NewID(request.Id), ports.AddItemToCartCommand{
		RestaurantID: vo.NewID(request.Body.RestaurantId),
		ProductID:    vo.NewID(request.Body.ProductId),
		Name:         request.Body.Name,
		Price:        price,
		Quantity:     request.Body.Quantity,
		Observation:  obs,
	})
	if err != nil {
		return nil, h.handleError(err)
	}
	return apigen.AddItemToCart200Response{}, nil
}

func (h *CustomerHandler) UpdateItemQuantity(ctx context.Context, request apigen.UpdateItemQuantityRequestObject) (apigen.UpdateItemQuantityResponseObject, error) {
	err := h.cartSvc.UpdateItemQuantity(ctx, vo.NewID(request.Id), ports.UpdateCartItemCommand{
		ProductID: vo.NewID(request.ProductId),
		Quantity:  request.Body.Quantity,
	})
	if err != nil {
		return nil, h.handleError(err)
	}
	return apigen.UpdateItemQuantity204Response{}, nil
}

func (h *CustomerHandler) RemoveItemFromCart(ctx context.Context, request apigen.RemoveItemFromCartRequestObject) (apigen.RemoveItemFromCartResponseObject, error) {
	err := h.cartSvc.RemoveItemFromCart(ctx, vo.NewID(request.Id), vo.NewID(request.ProductId))
	if err != nil {
		return nil, h.handleError(err)
	}
	return apigen.RemoveItemFromCart204Response{}, nil
}

func (h *CustomerHandler) ClearCart(ctx context.Context, request apigen.ClearCartRequestObject) (apigen.ClearCartResponseObject, error) {
	err := h.cartSvc.ClearCart(ctx, vo.NewID(request.Id))
	if err != nil {
		return nil, h.handleError(err)
	}
	return apigen.ClearCart204Response{}, nil
}

func (h *CustomerHandler) Checkout(ctx context.Context, request apigen.CheckoutRequestObject) (apigen.CheckoutResponseObject, error) {
	err := h.cartSvc.Checkout(ctx, vo.NewID(request.Id))
	if err != nil {
		return nil, h.handleError(err)
	}
	return apigen.Checkout202Response{}, nil
}

func (h *CustomerHandler) handleError(err error) error {
	return err
}

func ptr[T any](v T) *T {
	return &v
}
