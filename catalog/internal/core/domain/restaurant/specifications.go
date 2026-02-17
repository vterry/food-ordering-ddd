package restaurant

import (
	"errors"

	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	common "github.com/vterry/food-ordering/common/pkg"
)

var (
	ErrNameIsEmpty     = errors.New("name cannot be empty")
	ErrInvalidNameSize = errors.New("invalid name size - must have min: 5 and max: 30 characters")
)

const (
	NameMinSize = 5
	NameMaxSize = 30
)

type RestaurantParams struct {
	name    string
	address valueobjects.Address
}

func ValidateNewRestaurant(name string, address valueobjects.Address) error {
	params := RestaurantParams{
		name:    name,
		address: address,
	}
	spec := NewRestaurantSpecification()
	return spec(&params)
}

func NewRestaurantSpecification() common.Specification[RestaurantParams] {
	return common.And(
		ValidateNameSpec(),
		ValidateAddressSpec(),
	)
}

func ValidateNameSpec() common.Specification[RestaurantParams] {
	return func(k *RestaurantParams) error {
		if k.name == "" {
			return ErrNameIsEmpty
		}
		if len(k.name) < NameMinSize || len(k.name) > NameMaxSize {
			return ErrInvalidNameSize
		}
		return nil
	}
}

func ValidateAddressSpec() common.Specification[RestaurantParams] {
	return func(k *RestaurantParams) error {
		return k.address.Validate()
	}
}
