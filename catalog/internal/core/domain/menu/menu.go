package menu

import (
	"errors"

	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	common "github.com/vterry/food-ordering/common/pkg"
)

var (
	ErrMenuNotEditable     = errors.New("menu only can be edit in DRAFT status")
	ErrCategoryNotFound    = errors.New("category id not found in this menu")
	ErrCannotActivateEmpty = errors.New("cannot active an empty menu")
	ErrAlreadyActive       = errors.New("menu is already active")
	ErrAlreadyArchived     = errors.New("menu is already archived")
	ErrInvalidItemStatus   = errors.New("unknown item status")
	ErrMaxCategoryReached  = errors.New("max category reached in this menu")
)

const (
	MaxCategoriesPerMenu = 5
)

type Menu struct {
	valueobjects.MenuID
	name         string
	restaurantID valueobjects.RestaurantID
	status       enums.MenuStatus
	categories   []Category
	events       []common.DomainEvent
}

func NewMenu(name string, restaurantId valueobjects.RestaurantID) (*Menu, error) {
	if err := ValidateNewMenu(name, restaurantId); err != nil {
		return nil, err
	}

	menu := Menu{
		MenuID:       valueobjects.NewMenuID(),
		name:         name,
		restaurantID: restaurantId,
		status:       enums.MenuDraft,
		categories:   make([]Category, 0, MaxCategoriesPerMenu),
	}

	event := NewMenuCreated(menu)
	menu.AddEvent(event)
	return &menu, nil
}

func (m *Menu) Activate() error {
	if !m.hasItems() {
		return ErrCannotActivateEmpty
	}

	if m.status == enums.MenuActive {
		return ErrAlreadyActive
	}
	event := NewMenuActivated(*m)
	m.AddEvent(event)
	m.status = enums.MenuActive
	return nil
}

func (m *Menu) Archive() error {
	if m.status == enums.MenuArchived {
		return ErrAlreadyArchived
	}

	event := NewMenuArchived(m.MenuID)
	m.AddEvent(event)
	m.status = enums.MenuArchived
	return nil
}

func (m *Menu) AddCategory(category Category) error {
	if m.status != enums.MenuDraft {
		return ErrMenuNotEditable
	}

	if len(m.categories) >= MaxCategoriesPerMenu {
		return ErrMaxCategoryReached
	}

	m.categories = append(m.categories, category)

	event := NewMenuCategoryAdded(category)
	m.AddEvent(event)

	return nil
}

func (m *Menu) RemoveCategory(categoryId valueobjects.CategoryID) error {
	if m.status != enums.MenuDraft {
		return ErrMenuNotEditable
	}
	for i := range m.categories {
		if m.categories[i].CategoryID.Equals(categoryId) {
			m.categories = append(m.categories[:i], m.categories[i+1:]...)
			return nil
		}
	}
	return ErrCategoryNotFound
}

func (m *Menu) GetCategory(categoryId valueobjects.CategoryID) (*Category, error) {
	for i := range m.categories {
		if m.categories[i].CategoryID.Equals(categoryId) {
			return &m.categories[i], nil
		}
	}
	return nil, ErrCategoryNotFound
}

func (m *Menu) Categories() []Category {
	return m.categories
}

func (m *Menu) AddItemToCategory(categoryId valueobjects.CategoryID, item ItemMenu) error {
	if m.status != enums.MenuDraft {
		return ErrMenuNotEditable
	}

	for i := range m.categories {
		if m.categories[i].CategoryID.Equals(categoryId) {
			events, err := m.categories[i].AddItem(item)
			if err != nil {
				return err
			}

			m.AddEvent(events...)

			event := NewItemAddedToCategory(m.categories[i].CategoryID, item)
			m.AddEvent(event)
			return nil
		}
	}
	return ErrCategoryNotFound
}

func (m *Menu) FindItem(itemId valueobjects.ItemID) (*ItemMenu, bool) {
	for _, cat := range m.categories {
		if item, err := cat.GetItem(itemId); err == nil {
			return item, true
		}
	}
	return nil, false
}

func (m *Menu) RemoveItemFromCategory(categoryId valueobjects.CategoryID, item ItemMenu) error {
	if m.status != enums.MenuDraft {
		return ErrMenuNotEditable
	}

	category, err := m.GetCategory(categoryId)
	if err != nil {
		return err
	}

	return category.RemoveItem(item.ItemID)
}

func (m *Menu) UpdateItemPrice(categoryId valueobjects.CategoryID, itemId valueobjects.ItemID, price common.Money) error {
	category, err := m.GetCategory(categoryId)
	if err != nil {
		return err
	}

	itemRef, err := category.GetItem(itemId)
	if err != nil {
		return err
	}
	if err := itemRef.UpdatePrice(price); err != nil {
		return err
	}

	for _, event := range itemRef.PullEvent() {
		m.AddEvent(event)
	}

	return nil
}

func (m *Menu) UpdateItemAvailability(categoryId valueobjects.CategoryID, itemId valueobjects.ItemID, status enums.ItemStatus) error {
	category, err := m.GetCategory(categoryId)
	if err != nil {
		return err
	}

	itemRef, err := category.GetItem(itemId)
	if err != nil {
		return err
	}

	switch status {
	case enums.ItemAvailable:
		itemRef.MarkAvailable()
	case enums.ItemTempUnavailable:
		itemRef.MarkTemporarilyUnavailable()
	case enums.ItemUnavailable:
		itemRef.MarkUnavailable()
	default:
		return ErrInvalidItemStatus
	}

	for _, event := range itemRef.PullEvent() {
		m.AddEvent(event)
	}

	return nil
}

func (m *Menu) Name() string                            { return m.name }
func (m *Menu) RestaurantID() valueobjects.RestaurantID { return m.restaurantID }
func (m *Menu) Status() enums.MenuStatus                { return m.status }

func Restore(menuId valueobjects.MenuID,
	name string,
	restaurantId valueobjects.RestaurantID,
	status enums.MenuStatus,
	categories []Category) *Menu {
	return &Menu{
		MenuID:       menuId,
		name:         name,
		restaurantID: restaurantId,
		status:       status,
		categories:   categories,
		events:       []common.DomainEvent{},
	}
}

func (m *Menu) AddEvent(events ...common.DomainEvent) {
	m.events = append(m.events, events...)
}

func (m *Menu) PullEvent() []common.DomainEvent {
	events := m.events
	m.events = nil
	return events
}

func (m *Menu) hasItems() bool {
	for i := range m.categories {
		if !m.categories[i].IsEmpty() {
			return true
		}
	}
	return false
}
