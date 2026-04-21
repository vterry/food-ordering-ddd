package cart

import (
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"testing"
)

func TestCartInvariants(t *testing.T) {
	customerID := vo.NewID("cust-1")
	restaurant1 := vo.NewID("rest-1")
	restaurant2 := vo.NewID("rest-2")
	
	cartID := vo.NewID("cart-1")
	c := NewCart(cartID, customerID)

	money, _ := vo.NewMoney(10.0, "BRL")
	item := NewCartItem(vo.NewID("prod-1"), "Pizza", money, 1, "")

	// Add first item
	err := c.AddItem(restaurant1, item)
	if err != nil {
		t.Fatalf("unexpected error adding first item: %v", err)
	}

	// Test Event
	if len(c.Events()) != 1 {
		t.Errorf("expected 1 event, got %d", len(c.Events()))
	}

	// Add item from same restaurant
	err = c.AddItem(restaurant1, item)
	if err != nil {
		t.Errorf("unexpected error adding item from same restaurant: %v", err)
	}

	// Add item from different restaurant
	err = c.AddItem(restaurant2, item)
	if err != ErrDifferentRestaurant {
		t.Errorf("expected ErrDifferentRestaurant, got %v", err)
	}
}

func TestCartTotalValue(t *testing.T) {
	customerID := vo.NewID("cust-1")
	cart := NewCart(vo.NewID("cart-1"), customerID)
	restID := vo.NewID("rest-1")

	m10, _ := vo.NewMoney(10.0, "BRL")
	m20, _ := vo.NewMoney(20.0, "BRL")

	_ = cart.AddItem(restID, NewCartItem(vo.NewID("p1"), "Item 1", m10, 2, "")) // 20
	_ = cart.AddItem(restID, NewCartItem(vo.NewID("p2"), "Item 2", m20, 1, "")) // 20

	total := cart.TotalValue()
	if total.Amount() != 40.0 {
		t.Errorf("expected total 40.0, got %f", total.Amount())
	}
}
