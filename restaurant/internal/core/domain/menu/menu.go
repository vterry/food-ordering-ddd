package menu

import (
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

type MenuItem struct {
	id          vo.ID
	name        string
	description string
	price       vo.Money
	category    string
	isAvailable bool
}

func NewMenuItem(id vo.ID, name, description string, price vo.Money, category string) *MenuItem {
	return &MenuItem{
		id:          id,
		name:        name,
		description: description,
		price:       price,
		category:    category,
		isAvailable: true,
	}
}

func (i *MenuItem) ID() vo.ID {
	return i.id
}

func (i *MenuItem) Name() string {
	return i.name
}

func (i *MenuItem) Description() string {
	return i.description
}

func (i *MenuItem) Price() vo.Money {
	return i.price
}

func (i *MenuItem) Category() string {
	return i.category
}

func (i *MenuItem) IsAvailable() bool {
	return i.isAvailable
}

func (i *MenuItem) SetAvailability(available bool) {
	i.isAvailable = available
}

type Menu struct {
	base.BaseAggregateRoot
	restaurantID vo.ID
	name         string
	isActive     bool
	items        []*MenuItem
}

func NewMenu(id, restaurantID vo.ID, name string) *Menu {
	m := &Menu{
		restaurantID: restaurantID,
		name:         name,
		isActive:     false,
		items:        []*MenuItem{},
	}
	m.SetID(id)
	return m
}

func (m *Menu) RestaurantID() vo.ID {
	return m.restaurantID
}

func (m *Menu) Name() string {
	return m.name
}

func (m *Menu) IsActive() bool {
	return m.isActive
}

func (m *Menu) Items() []*MenuItem {
	return m.items
}

func (m *Menu) Activate() {
	m.isActive = true
	m.AddEvent(NewMenuActivatedEvent(m.ID(), m.restaurantID))
}

func (m *Menu) Deactivate() {
	m.isActive = false
}

func (m *Menu) AddItem(item *MenuItem) {
	m.items = append(m.items, item)
}

func (m *Menu) ChangeItemAvailability(productID vo.ID, available bool) {
	for _, item := range m.items {
		if item.id.Equals(productID) {
			item.SetAvailability(available)
			m.AddEvent(NewItemAvailabilityChangedEvent(m.ID(), productID, available))
			break
		}
	}
}
