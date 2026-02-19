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

	m.status = enums.MenuActive
	event := NewMenuActivated(*m)
	m.AddEvent(event)

	return nil
}

func (m *Menu) Archive() error {
	if m.status == enums.MenuArchived {
		return ErrAlreadyArchived
	}

	m.status = enums.MenuArchived

	event := NewMenuArchived(m.MenuID)
	m.AddEvent(event)

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
			removed := m.categories[i]
			m.categories = append(m.categories[:i], m.categories[i+1:]...)
			event := NewMenuCategoryRemoved(removed)
			m.AddEvent(event)
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
			err := m.categories[i].AddItem(item)
			if err != nil {
				return err
			}

			event := NewItemMenuCreated(m.categories[i].CategoryID, item)
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

	if err := category.RemoveItem(item.ItemID); err != nil {
		return err
	}

	event := NewItemMenuRemoved(categoryId, item)
	m.AddEvent(event)

	return nil
}

func (m *Menu) UpdateItemName(categoryId valueobjects.CategoryID, itemId valueobjects.ItemID, newName string) error {
	category, err := m.GetCategory(categoryId)
	if err != nil {
		return err
	}

	itemRef, err := category.GetItem(itemId)
	if err != nil {
		return err
	}

	oldName := itemRef.Name()

	if err := itemRef.UpdateName(newName); err != nil {
		return err
	}

	event := NewItemMenuNameChanged(categoryId, itemId, oldName, newName)
	m.AddEvent(event)

	return nil
}

func (m *Menu) UpdateItemDescription(categoryId valueobjects.CategoryID, itemId valueobjects.ItemID, newDesc string) error {
	category, err := m.GetCategory(categoryId)
	if err != nil {
		return err
	}

	itemRef, err := category.GetItem(itemId)
	if err != nil {
		return err
	}

	if err := itemRef.UpdateDescription(newDesc); err != nil {
		return err
	}

	return nil
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

	oldPrice := itemRef.BasePrice()

	if err := itemRef.UpdatePrice(price); err != nil {
		return err
	}

	event := NewItemMenuPriceChanged(categoryId, itemId, oldPrice, price)
	m.AddEvent(event)

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

	oldStatus := itemRef.Status()

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

	event := NewItemMenuAvailabilityChanged(categoryId, itemId, oldStatus, status)
	m.AddEvent(event)

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

// ItemValidationResult holds the outcome of validating a single item in the menu.
type ItemValidationResult struct {
	Item  *ItemMenu // non-nil when validation passes
	Error string    // non-empty when validation fails
}

// ValidateItems checks each itemID against the active menu: existence and availability.
// Parsing of external IDs is the caller's responsibility.
func (m *Menu) ValidateItems(itemIDs []valueobjects.ItemID) []ItemValidationResult {
	results := make([]ItemValidationResult, 0, len(itemIDs))
	for _, id := range itemIDs {
		item, found := m.FindItem(id)
		if !found {
			results = append(results, ItemValidationResult{Error: "item " + id.String() + " not found in active menu"})
			continue
		}
		if item.Status() != enums.ItemAvailable {
			results = append(results, ItemValidationResult{Error: "item " + item.Name() + " is not available"})
			continue
		}
		results = append(results, ItemValidationResult{Item: item})
	}
	return results
}

func (m *Menu) hasItems() bool {
	for i := range m.categories {
		if !m.categories[i].IsEmpty() {
			return true
		}
	}
	return false
}
