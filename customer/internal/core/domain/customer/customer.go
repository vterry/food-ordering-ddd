package customer

import (
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

type Customer struct {
	base.BaseAggregateRoot
	name      Name
	email     Email
	phone     Phone
	addresses []*Address
}

func NewCustomer(id vo.ID, name Name, email Email, phone Phone) *Customer {
	c := &Customer{
		name:      name,
		email:     email,
		phone:     phone,
		addresses: []*Address{},
	}
	c.SetID(id)

	// Iniciar com evento de registro
	c.AddEvent(NewCustomerRegisteredEvent(id, name.String(), email.String()))

	return c
}

func (c *Customer) Name() Name {
	return c.name
}

func (c *Customer) Email() Email {
	return c.email
}

func (c *Customer) Phone() Phone {
	return c.phone
}

func (c *Customer) Addresses() []*Address {
	return c.addresses
}

func (c *Customer) AddAddress(address *Address) {
	// Regra de negócio: se for o primeiro endereço, torna-o padrão automaticamente
	if len(c.addresses) == 0 {
		address.isDefault = true
	}
	c.addresses = append(c.addresses, address)
}

func (c *Customer) ChangeName(newName Name) {
	c.name = newName
	// Poderia adicionar um evento CustomerNameChanged aqui
}

func (c *Customer) SetDefaultAddress(addressID vo.ID) {
	for _, addr := range c.addresses {
		addr.isDefault = addr.ID().Equals(addressID)
	}
}
