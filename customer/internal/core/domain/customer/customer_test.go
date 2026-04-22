package customer

import (
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"testing"
)

func TestNewCustomer(t *testing.T) {
	id := vo.NewID("123")
	name, _ := NewName("John Doe")
	email, _ := NewEmail("john@example.com")
	phone, _ := NewPhone("5511999999999")

	tests := []struct {
		name     string
		id       vo.ID
		custName Name
		email    Email
		phone    Phone
	}{
		{
			name:     "valid customer creation",
			id:       id,
			custName: name,
			email:    email,
			phone:    phone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCustomer(tt.id, tt.custName, tt.email, tt.phone)

			if !c.ID().Equals(tt.id) {
				t.Errorf("expected ID %v, got %v", tt.id, c.ID())
			}
			if c.Name().String() != tt.custName.String() {
				t.Errorf("expected name %s, got %s", tt.custName.String(), c.Name().String())
			}

			// Testar registro de eventos
			if len(c.Events()) != 1 {
				t.Errorf("expected 1 event, got %d", len(c.Events()))
			}
			event, ok := c.Events()[0].(CustomerRegisteredEvent)
			if !ok {
				t.Errorf("expected CustomerRegisteredEvent, got %T", c.Events()[0])
			}
			if event.Name != tt.custName.String() {
				t.Errorf("expected event name %s, got %s", tt.custName.String(), event.Name)
			}
		})
	}
}

func TestCustomer_AddressLogic(t *testing.T) {
	customerID := vo.NewID("123")
	name, _ := NewName("John Doe")
	email, _ := NewEmail("john@example.com")
	phone, _ := NewPhone("5511999999999")

	t.Run("first address becomes default automatically", func(t *testing.T) {
		c := NewCustomer(customerID, name, email, phone)
		addr1 := NewAddress(vo.NewID("addr1"), customerID, "Street 1", "City", "12345", false)

		c.AddAddress(addr1)
		if !addr1.IsDefault() {
			t.Errorf("expected first address to be default")
		}
	})

	t.Run("secondary address is not default by default", func(t *testing.T) {
		c := NewCustomer(customerID, name, email, phone)
		addr1 := NewAddress(vo.NewID("addr1"), customerID, "Street 1", "City", "12345", false)
		addr2 := NewAddress(vo.NewID("addr2"), customerID, "Street 2", "City", "12345", false)

		c.AddAddress(addr1)
		c.AddAddress(addr2)

		if !addr1.IsDefault() {
			t.Errorf("addr1 should be default")
		}
		if addr2.IsDefault() {
			t.Errorf("addr2 should not be default")
		}
	})

	t.Run("changing default address", func(t *testing.T) {
		c := NewCustomer(customerID, name, email, phone)
		addr1 := NewAddress(vo.NewID("addr1"), customerID, "Street 1", "City", "12345", false)
		addr2 := NewAddress(vo.NewID("addr2"), customerID, "Street 2", "City", "12345", false)

		c.AddAddress(addr1)
		c.AddAddress(addr2)

		c.SetDefaultAddress(addr2.ID())
		if !addr2.IsDefault() {
			t.Errorf("addr2 should be default after SetDefaultAddress")
		}
		if addr1.IsDefault() {
			t.Errorf("addr1 should no longer be default")
		}
	})
}

func TestCustomer_ChangeName(t *testing.T) {
	id := vo.NewID("123")
	name, _ := NewName("John Doe")
	email, _ := NewEmail("john@example.com")
	phone, _ := NewPhone("5511999999999")
	c := NewCustomer(id, name, email, phone)

	newName, _ := NewName("Jane Doe")
	c.ChangeName(newName)

	if c.Name().String() != "Jane Doe" {
		t.Errorf("expected name Jane Doe, got %s", c.Name().String())
	}
}

func TestCustomer_Getters(t *testing.T) {
	id := vo.NewID("123")
	name, _ := NewName("John Doe")
	email, _ := NewEmail("john@example.com")
	phone, _ := NewPhone("5511999999999")
	c := NewCustomer(id, name, email, phone)

	if c.Email().String() != "john@example.com" {
		t.Errorf("expected email john@example.com, got %s", c.Email().String())
	}
	if c.Phone().String() != "5511999999999" {
		t.Errorf("expected phone 5511999999999, got %s", c.Phone().String())
	}
	if len(c.Addresses()) != 0 {
		t.Errorf("expected 0 addresses, got %d", len(c.Addresses()))
	}
}

func TestCustomerEvents_EventType(t *testing.T) {
	id := vo.NewID("123")
	ev := NewCustomerRegisteredEvent(id, "John", "john@ex.com")
	if ev.EventType() != "CustomerRegistered" {
		t.Errorf("expected EventType CustomerRegistered, got %s", ev.EventType())
	}
}


