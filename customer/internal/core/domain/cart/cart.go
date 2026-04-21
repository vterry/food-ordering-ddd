package cart

import (
	"errors"
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

var (
	ErrDifferentRestaurant = errors.New("cannot add item from a different restaurant")
)

type CartItem struct {
	productID   vo.ID
	name        string
	price       vo.Money
	quantity    int
	observation string
}

func NewCartItem(id vo.ID, name string, price vo.Money, quantity int, observation string) CartItem {
	return CartItem{
		productID:   id,
		name:        name,
		price:       price,
		quantity:    quantity,
		observation: observation,
	}
}

func (i CartItem) ProductID() vo.ID {
	return i.productID
}

func (i CartItem) Name() string {
	return i.name
}

func (i CartItem) Price() vo.Money {
	return i.price
}

func (i CartItem) Quantity() int {
	return i.quantity
}

func (i CartItem) Observation() string {
	return i.observation
}

type Cart struct {
	base.BaseAggregateRoot
	customerID   vo.ID
	restaurantID vo.ID
	items        []CartItem
}

func NewCart(id vo.ID, customerID vo.ID) *Cart {
	c := &Cart{
		customerID: customerID,
	}
	c.SetID(id)
	return c
}

func (c *Cart) CustomerID() vo.ID {
	return c.customerID
}

func (c *Cart) RestaurantID() vo.ID {
	return c.restaurantID
}

func (c *Cart) Items() []CartItem {
	return c.items
}

func (c *Cart) AddItem(restaurantID vo.ID, item CartItem) error {
	if !c.restaurantID.IsEmpty() && !c.restaurantID.Equals(restaurantID) {
		return ErrDifferentRestaurant
	}

	if c.restaurantID.IsEmpty() {
		c.restaurantID = restaurantID
	}

	// Se o item já existe, atualiza quantidade
	for i, existing := range c.items {
		if existing.productID.Equals(item.productID) {
			c.items[i].quantity += item.quantity
			return nil
		}
	}

	c.items = append(c.items, item)

	// Produzir evento de domínio
	c.AddEvent(NewItemAddedToCartEvent(c.customerID, restaurantID, item.productID, item.quantity))

	return nil
}

func (c *Cart) RemoveItem(productID vo.ID) {
	for i, item := range c.items {
		if item.productID.Equals(productID) {
			c.items = append(c.items[:i], c.items[i+1:]...)
			break
		}
	}
	if len(c.items) == 0 {
		c.restaurantID = vo.ID{}
	}
}

func (c *Cart) UpdateItemQuantity(productID vo.ID, quantity int) {
	if quantity <= 0 {
		c.RemoveItem(productID)
		return
	}
	for i, item := range c.items {
		if item.productID.Equals(productID) {
			c.items[i].quantity = quantity
			break
		}
	}
}

func (c *Cart) Clear() {
	c.items = nil
	c.restaurantID = vo.ID{}
	// Poderia adicionar um evento CartCleared aqui
}

func (c *Cart) Checkout() error {
	if len(c.items) == 0 {
		return errors.New("cannot checkout an empty cart")
	}
	
	c.AddEvent(NewCheckoutRequestedEvent(c.customerID, c.restaurantID, c.items))
	return nil
}

func (c *Cart) TotalValue() vo.Money {
	if len(c.items) == 0 {
		m, _ := vo.NewMoney(0, "BRL")
		return m
	}

	total, _ := vo.NewMoney(0, c.items[0].price.Currency())
	for _, item := range c.items {
		val, _ := item.price.Multiply(float64(item.quantity))
		total, _ = total.Add(val)
	}
	return total
}
