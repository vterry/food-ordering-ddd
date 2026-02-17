package menu

import (
	"errors"

	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	common "github.com/vterry/food-ordering/common/pkg"
)

var (
	ErrItemNameIsEmpty        = errors.New("item name cannot be empty")
	ErrInvalidItemNameSize    = errors.New("invalid item name size - must have min: 3 and max: 15 characters")
	ErrItemPriceInvalid       = errors.New("item price must be greater than zero")
	ErrItemDescriptionIsEmpty = errors.New("item description cannot be empty")
	ErrCategoryNameIsEmpty    = errors.New("category name cannot be empty")
	ErrMenuNameIsEmpty        = errors.New("menu name cannot be empty")
	ErrInvalidMenuNameSize    = errors.New("invalid menu name size - must have min: 5 and max: 20 characters")
	ErrRestaurantIdIsNil      = errors.New("restaurant id is nil")
)

const (
	MinLengthMenuNameSize = 5
	MaxLengthMenuNameSize = 20
	MinLengthItemNameSize = 3
	MaxLengthItemNameSize = 15
)

// Menu Specs
type MenuParams struct {
	name         string
	restaurantID valueobjects.RestaurantID
}

func ValidateNewMenu(name string, restaurantId valueobjects.RestaurantID) error {
	params := MenuParams{
		name:         name,
		restaurantID: restaurantId,
	}
	spec := NewMenuSpecification()
	return spec(&params)
}

func NewMenuSpecification() common.Specification[MenuParams] {
	return common.And(
		MenuNameSpec(),
		RestaurantIDSpec(),
	)
}

func MenuNameSpec() common.Specification[MenuParams] {
	return func(m *MenuParams) error {
		return ValidateMenuName(m.name)
	}
}

func RestaurantIDSpec() common.Specification[MenuParams] {
	return func(m *MenuParams) error {
		return ValidateRestaurantID(m.restaurantID)
	}
}

func ValidateMenuName(name string) error {
	if name == "" {
		return ErrMenuNameIsEmpty
	}
	if len(name) < MinLengthMenuNameSize || len(name) > MaxLengthMenuNameSize {
		return ErrInvalidMenuNameSize
	}
	return nil
}

func ValidateRestaurantID(restaurantID valueobjects.RestaurantID) error {
	if restaurantID.IsZero() {
		return ErrRestaurantIdIsNil
	}
	return nil
}

// Category Specs
type CategoryParams struct {
	name string
}

func ValidateNewCategory(name string) error {
	params := CategoryParams{
		name: name,
	}

	spec := NewCategorySpecification()
	return spec(&params)
}

func NewCategorySpecification() common.Specification[CategoryParams] {
	return common.And(
		CategoryNameSpec(),
	)
}
func CategoryNameSpec() common.Specification[CategoryParams] {
	return func(c *CategoryParams) error {
		return ValidateCategoryName(c.name)
	}
}

func ValidateCategoryName(name string) error {
	if name == "" {
		return ErrCategoryNameIsEmpty
	}
	return nil
}

// Item Menu Specs
type ItemParams struct {
	name        string
	description string
	price       common.Money
}

func ValidateNewItemMenu(name, description string, price common.Money) error {
	params := ItemParams{
		name:        name,
		description: description,
		price:       price,
	}
	spec := NewItemMenuSpecification()
	return spec(&params)
}

func NewItemMenuSpecification() common.Specification[ItemParams] {
	return common.And(
		ItemNameSpec(),
		ItemPriceSpec(),
		ItemDescriptionSpec(),
	)
}

func ItemNameSpec() common.Specification[ItemParams] {
	return func(i *ItemParams) error {
		return ValidateItemName(i.name)
	}
}

func ItemPriceSpec() common.Specification[ItemParams] {
	return func(i *ItemParams) error {
		return ValidateItemPrice(i.price)
	}
}

func ItemDescriptionSpec() common.Specification[ItemParams] {
	return func(i *ItemParams) error {
		return ValidateItemDescription(i.description)
	}
}

func ValidateItemName(name string) error {
	if name == "" {
		return ErrItemNameIsEmpty
	}
	if len(name) < MinLengthItemNameSize || len(name) > MaxLengthItemNameSize {
		return ErrInvalidItemNameSize
	}
	return nil
}

func ValidateItemPrice(price common.Money) error {
	if !price.IsGreaterThanZero() {
		return ErrItemPriceInvalid
	}
	return nil
}

func ValidateItemDescription(description string) error {
	if description == "" {
		return ErrItemDescriptionIsEmpty
	}
	return nil
}
