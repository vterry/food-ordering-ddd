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

	c := NewCustomer(id, name, email, phone)

	if !c.ID().Equals(id) {
		t.Errorf("expected ID %v, got %v", id, c.ID())
	}
	if c.Name().String() != "John Doe" {
		t.Errorf("expected name John Doe, got %s", c.Name().String())
	}

	// Testar registro de eventos
	if len(c.Events()) != 1 {
		t.Errorf("expected 1 event, got %d", len(c.Events()))
	}
	event := c.Events()[0].(CustomerRegisteredEvent)
	if event.Name != "John Doe" {
		t.Errorf("expected event name John Doe, got %s", event.Name)
	}
}

func TestCustomerAddressDefaultLogic(t *testing.T) {
	id := vo.NewID("123")
	name, _ := NewName("John Doe")
	email, _ := NewEmail("john@example.com")
	phone, _ := NewPhone("5511999999999")
	c := NewCustomer(id, name, email, phone)

	addr1 := NewAddress(vo.NewID("addr1"), id, "Street 1", "City", "12345", false)
	addr2 := NewAddress(vo.NewID("addr2"), id, "Street 2", "City", "12345", false)

	c.AddAddress(addr1)
	if !addr1.IsDefault() {
		t.Errorf("first address should be default")
	}

	c.AddAddress(addr2)
	if addr2.IsDefault() {
		t.Errorf("second address should not be default by default")
	}

	c.SetDefaultAddress(addr2.ID())
	if !addr2.IsDefault() {
		t.Errorf("addr2 should be default after SetDefaultAddress")
	}
	if addr1.IsDefault() {
		t.Errorf("addr1 should no longer be default")
	}
}
