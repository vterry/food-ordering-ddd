package order

import (
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

// OrderItem is a value object representing an item in an order.
type OrderItem struct {
	menuItemID vo.ID
	name       string
	quantity   int
	unitPrice  vo.Money
	notes      string
}

// NewOrderItem creates a new OrderItem.
func NewOrderItem(menuItemID vo.ID, name string, quantity int, unitPrice vo.Money, notes string) OrderItem {
	return OrderItem{
		menuItemID: menuItemID,
		name:       name,
		quantity:   quantity,
		unitPrice:  unitPrice,
		notes:      notes,
	}
}

func (i OrderItem) MenuItemID() vo.ID {
	return i.menuItemID
}

func (i OrderItem) Name() string {
	return i.name
}

func (i OrderItem) Quantity() int {
	return i.quantity
}

func (i OrderItem) UnitPrice() vo.Money {
	return i.unitPrice
}

func (i OrderItem) Notes() string {
	return i.notes
}

func (i OrderItem) TotalPrice() vo.Money {
	total, _ := i.unitPrice.Multiply(float64(i.quantity))
	return total
}
