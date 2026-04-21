package customer

import (
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

type Address struct {
	base.BaseAggregateRoot
	customerID vo.ID
	street     string
	city       string
	zipCode    string
	isDefault  bool
}

func NewAddress(id vo.ID, customerID vo.ID, street, city, zipCode string, isDefault bool) *Address {
	a := &Address{
		customerID: customerID,
		street:     street,
		city:       city,
		zipCode:    zipCode,
		isDefault:  isDefault,
	}
	a.SetID(id)
	return a
}

func (a *Address) CustomerID() vo.ID {
	return a.customerID
}

func (a *Address) Street() string {
	return a.street
}

func (a *Address) City() string {
	return a.city
}

func (a *Address) ZipCode() string {
	return a.zipCode
}

func (a *Address) IsDefault() bool {
	return a.isDefault
}
