package valueobjects

import (
	"errors"

	common "github.com/vterry/food-ordering/common/pkg"
)

var (
	ErrProductIdIsEmpty      = errors.New("product id cannot be empty")
	ErrProductNameIsEmpty    = errors.New("product name cannot be empty")
	ErrQntLessThanZero       = errors.New("quantity cannot be less than zero")
	ErrUnitPriceLessThanZero = errors.New("unit price cannot be less than zero")
	ErrInvalidTotalPrice     = errors.New("incorrect total price")
)

type OrderItem struct {
	productId   string
	productName string
	unitPrice   common.Money
	quantity    int
	totalPrice  common.Money
}

func NewOrderItem(productId, productName string, unitPrice common.Money, quantity int) (OrderItem, error) {

	totalPrice := unitPrice.Amount() * int64(quantity)

	orderItem := OrderItem{
		productId:   productId,
		productName: productName,
		unitPrice:   unitPrice,
		quantity:    quantity,
		totalPrice:  common.NewMoneyFromCents(totalPrice),
	}

	if err := orderItem.Validate(); err != nil {
		return OrderItem{}, err
	}
	return orderItem, nil
}

func (o *OrderItem) Validate() error {

	if o.productId == "" {
		return ErrProductIdIsEmpty
	}

	if o.productName == "" {
		return ErrProductNameIsEmpty
	}

	if !o.unitPrice.IsGreaterThanZero() {
		return ErrUnitPriceLessThanZero
	}

	if o.quantity <= 0 {
		return ErrQntLessThanZero
	}

	return nil
}

func (o *OrderItem) TotalPrice() common.Money {
	return o.totalPrice
}
