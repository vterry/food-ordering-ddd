package menu

import (
	"errors"

	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
)

var (
	ErrItemNotInCategory = errors.New("item not found in this category")
	ErrItemAlreadyExist  = errors.New("item already exists in this category")
	ErrMaxItemsReached   = errors.New("this category reach max number of items")
)

const (
	MaxItemPerCategory = 30
)

type Category struct {
	valueobjects.CategoryID
	name  string
	items []ItemMenu
}

func NewCategory(name string) (*Category, error) {
	if err := ValidateNewCategory(name); err != nil {
		return nil, err
	}

	return &Category{
		CategoryID: valueobjects.NewCategoryID(),
		name:       name,
		items:      make([]ItemMenu, 0, MaxItemPerCategory),
	}, nil
}

func (c *Category) AddItem(item ItemMenu) error {
	for i := range c.items {
		if c.items[i].ItemID.Equals(item.ItemID) {
			return ErrItemAlreadyExist
		}
	}

	if len(c.items) >= MaxItemPerCategory {
		return ErrMaxItemsReached
	}

	c.items = append(c.items, item)

	return nil
}

func (c *Category) RemoveItem(itemId valueobjects.ItemID) error {
	for i := range c.items {
		if c.items[i].ItemID.Equals(itemId) {
			c.items = append(c.items[:i], c.items[i+1:]...)
			return nil
		}
	}
	return ErrItemNotInCategory
}

func (c *Category) GetItem(itemId valueobjects.ItemID) (*ItemMenu, error) {
	for i := range c.items {
		if c.items[i].ItemID.Equals(itemId) {
			return &c.items[i], nil
		}
	}
	return nil, ErrItemNotInCategory
}

func (c *Category) IsEmpty() bool {
	return len(c.items) == 0
}

func (c *Category) Rename(name string) error {
	if err := ValidateCategoryName(name); err != nil {
		return err
	}
	c.name = name
	return nil
}

func (c *Category) Name() string      { return c.name }
func (c *Category) Items() []ItemMenu { return c.items }

func RestoreCategory(categoryId valueobjects.CategoryID, name string, items []ItemMenu) *Category {
	return &Category{
		CategoryID: categoryId,
		name:       name,
		items:      items,
	}
}
