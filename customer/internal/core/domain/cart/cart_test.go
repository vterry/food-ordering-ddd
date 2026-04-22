package cart

import (
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"testing"
)

func TestCart_AddItem(t *testing.T) {
	customerID := vo.NewID("cust-1")
	restID := vo.NewID("rest-1")
	money, _ := vo.NewMoney(10.0, "BRL")

	t.Run("add new item emits event", func(t *testing.T) {
		c := NewCart(vo.NewID("cart-1"), customerID)
		item := NewCartItem(vo.NewID("p1"), "Pizza", money, 1, "no onions")

		err := c.AddItem(restID, item)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(c.Items()) != 1 {
			t.Errorf("expected 1 item, got %d", len(c.Items()))
		}
		if c.Items()[0].Observation() != "no onions" {
			t.Errorf("expected observation 'no onions', got '%s'", c.Items()[0].Observation())
		}
		if len(c.Events()) != 1 {
			t.Errorf("expected 1 event, got %d", len(c.Events()))
		}
	})

	t.Run("add duplicate item merges quantity and does not emit new event", func(t *testing.T) {
		c := NewCart(vo.NewID("cart-1"), customerID)
		item := NewCartItem(vo.NewID("p1"), "Pizza", money, 1, "")

		_ = c.AddItem(restID, item)
		c.ClearEvents()

		err := c.AddItem(restID, item)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(c.Items()) != 1 {
			t.Errorf("expected 1 item, got %d", len(c.Items()))
		}
		if c.Items()[0].Quantity() != 2 {
			t.Errorf("expected quantity 2, got %d", c.Items()[0].Quantity())
		}
		if len(c.Events()) != 0 {
			t.Errorf("expected 0 events, got %d", len(c.Events()))
		}
	})

	t.Run("add item from different restaurant fails", func(t *testing.T) {
		c := NewCart(vo.NewID("cart-1"), customerID)
		_ = c.AddItem(restID, NewCartItem(vo.NewID("p1"), "Pizza", money, 1, ""))

		err := c.AddItem(vo.NewID("rest-2"), NewCartItem(vo.NewID("p2"), "Burger", money, 1, ""))
		if err != ErrDifferentRestaurant {
			t.Errorf("expected ErrDifferentRestaurant, got %v", err)
		}
	})
}

func TestCart_RemoveItem(t *testing.T) {
	customerID := vo.NewID("cust-1")
	restID := vo.NewID("rest-1")
	money, _ := vo.NewMoney(10.0, "BRL")

	t.Run("remove item reduces count", func(t *testing.T) {
		c := NewCart(vo.NewID("cart-1"), customerID)
		p1 := vo.NewID("p1")
		_ = c.AddItem(restID, NewCartItem(p1, "Pizza", money, 1, ""))
		_ = c.AddItem(restID, NewCartItem(vo.NewID("p2"), "Burger", money, 1, ""))

		c.RemoveItem(p1)
		if len(c.Items()) != 1 {
			t.Errorf("expected 1 item, got %d", len(c.Items()))
		}
	})

	t.Run("removing last item resets restaurant ID", func(t *testing.T) {
		c := NewCart(vo.NewID("cart-1"), customerID)
		p1 := vo.NewID("p1")
		_ = c.AddItem(restID, NewCartItem(p1, "Pizza", money, 1, ""))

		c.RemoveItem(p1)
		if !c.RestaurantID().IsEmpty() {
			t.Errorf("expected empty restaurant ID, got %v", c.RestaurantID())
		}
	})
}

func TestCart_UpdateQuantity(t *testing.T) {
	customerID := vo.NewID("cust-1")
	restID := vo.NewID("rest-1")
	money, _ := vo.NewMoney(10.0, "BRL")

	t.Run("update quantity to valid value", func(t *testing.T) {
		c := NewCart(vo.NewID("cart-1"), customerID)
		p1 := vo.NewID("p1")
		_ = c.AddItem(restID, NewCartItem(p1, "Pizza", money, 1, ""))

		c.UpdateItemQuantity(p1, 5)
		if c.Items()[0].Quantity() != 5 {
			t.Errorf("expected quantity 5, got %d", c.Items()[0].Quantity())
		}
	})

	t.Run("update quantity to zero removes item", func(t *testing.T) {
		c := NewCart(vo.NewID("cart-1"), customerID)
		p1 := vo.NewID("p1")
		_ = c.AddItem(restID, NewCartItem(p1, "Pizza", money, 1, ""))

		c.UpdateItemQuantity(p1, 0)
		if len(c.Items()) != 0 {
			t.Errorf("expected 0 items, got %d", len(c.Items()))
		}
	})
}

func TestCart_Clear(t *testing.T) {
	customerID := vo.NewID("cust-1")
	restID := vo.NewID("rest-1")
	money, _ := vo.NewMoney(10.0, "BRL")

	c := NewCart(vo.NewID("cart-1"), customerID)
	_ = c.AddItem(restID, NewCartItem(vo.NewID("p1"), "Pizza", money, 1, ""))

	c.Clear()
	if len(c.Items()) != 0 {
		t.Errorf("expected 0 items after clear")
	}
	if !c.RestaurantID().IsEmpty() {
		t.Errorf("expected empty restaurant ID after clear")
	}
}

func TestCart_Checkout(t *testing.T) {
	customerID := vo.NewID("cust-1")
	restID := vo.NewID("rest-1")
	money, _ := vo.NewMoney(10.0, "BRL")

	t.Run("checkout with items emits event", func(t *testing.T) {
		c := NewCart(vo.NewID("cart-1"), customerID)
		_ = c.AddItem(restID, NewCartItem(vo.NewID("p1"), "Pizza", money, 1, ""))
		c.ClearEvents()

		err := c.Checkout()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(c.Events()) != 1 {
			t.Errorf("expected 1 event, got %d", len(c.Events()))
		}
		_, ok := c.Events()[0].(CheckoutRequestedEvent)
		if !ok {
			t.Errorf("expected CheckoutRequestedEvent, got %T", c.Events()[0])
		}
	})

	t.Run("checkout empty cart fails", func(t *testing.T) {
		c := NewCart(vo.NewID("cart-1"), customerID)
		err := c.Checkout()
		if err == nil {
			t.Error("expected error checking out empty cart")
		}
	})
}

func TestCart_TotalValue(t *testing.T) {
	customerID := vo.NewID("cust-1")

	t.Run("total value of empty cart is zero", func(t *testing.T) {
		c := NewCart(vo.NewID("cart-1"), customerID)
		total := c.TotalValue()
		if total.Amount() != 0 {
			t.Errorf("expected total 0, got %f", total.Amount())
		}
	})

	t.Run("total value with items", func(t *testing.T) {
		c := NewCart(vo.NewID("cart-1"), customerID)
		restID := vo.NewID("rest-1")
		m10, _ := vo.NewMoney(10.0, "BRL")
		m20, _ := vo.NewMoney(20.0, "BRL")

		_ = c.AddItem(restID, NewCartItem(vo.NewID("p1"), "Item 1", m10, 2, "")) // 20
		_ = c.AddItem(restID, NewCartItem(vo.NewID("p2"), "Item 2", m20, 1, "")) // 20

		total := c.TotalValue()
		if total.Amount() != 40.0 {
			t.Errorf("expected total 40.0, got %f", total.Amount())
		}
	})
}

func TestCartItem_Getters(t *testing.T) {
	id := vo.NewID("p1")
	name := "Pizza"
	price, _ := vo.NewMoney(10.0, "BRL")
	qty := 2
	obs := "no olives"

	item := NewCartItem(id, name, price, qty, obs)

	if !item.ProductID().Equals(id) {
		t.Errorf("expected ProductID %v, got %v", id, item.ProductID())
	}
	if item.Name() != name {
		t.Errorf("expected name %s, got %s", name, item.Name())
	}
	if item.Price().Amount() != price.Amount() {
		t.Errorf("expected price %f, got %f", price.Amount(), item.Price().Amount())
	}
	if item.Quantity() != qty {
		t.Errorf("expected qty %d, got %d", qty, item.Quantity())
	}
	if item.Observation() != obs {
		t.Errorf("expected obs %s, got %s", obs, item.Observation())
	}
}

func TestCart_Getters(t *testing.T) {
	customerID := vo.NewID("cust1")
	c := NewCart(vo.NewID("cart1"), customerID)

	if !c.CustomerID().Equals(customerID) {
		t.Errorf("expected CustomerID %v, got %v", customerID, c.CustomerID())
	}
}

func TestCartEvents_EventType(t *testing.T) {
	customerID := vo.NewID("cust1")
	restID := vo.NewID("rest1")
	prodID := vo.NewID("prod1")

	t.Run("ItemAddedToCartEvent", func(t *testing.T) {
		ev := NewItemAddedToCartEvent(customerID, restID, prodID, 1)
		if ev.EventType() != "ItemAddedToCart" {
			t.Errorf("expected EventType ItemAddedToCart, got %s", ev.EventType())
		}
	})

	t.Run("CheckoutRequestedEvent", func(t *testing.T) {
		ev := NewCheckoutRequestedEvent(customerID, restID, nil)
		if ev.EventType() != "cart.checkout_requested" {
			t.Errorf("expected EventType cart.checkout_requested, got %s", ev.EventType())
		}
	})
}


