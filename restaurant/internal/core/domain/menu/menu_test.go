package menu

import (
	"testing"

	"github.com/vterry/food-project/common/pkg/domain/vo"
)

func TestNewMenu(t *testing.T) {
	id := vo.NewID("menu-1")
	restaurantID := vo.NewID("rest-1")
	name := "Main Menu"

	m := NewMenu(id, restaurantID, name)

	if !m.ID().Equals(id) {
		t.Errorf("expected ID %v, got %v", id, m.ID())
	}
	if !m.RestaurantID().Equals(restaurantID) {
		t.Errorf("expected RestaurantID %v, got %v", restaurantID, m.RestaurantID())
	}
	if m.Name() != name {
		t.Errorf("expected name %s, got %s", name, m.Name())
	}
	if m.IsActive() {
		t.Errorf("new menu should be inactive by default")
	}
}

func TestMenu_Activate(t *testing.T) {
	m := NewMenu(vo.NewID("m-1"), vo.NewID("r-1"), "Menu")
	m.Activate()
	if !m.IsActive() {
		t.Errorf("menu should be active after calling Activate()")
	}
	if len(m.Events()) != 1 {
		t.Errorf("expected 1 event, got %d", len(m.Events()))
	}
	if m.Events()[0].EventType() != "restaurant.menu.activated" {
		t.Errorf("expected event type restaurant.menu.activated, got %s", m.Events()[0].EventType())
	}
}

func TestMenu_AddItem(t *testing.T) {
	m := NewMenu(vo.NewID("m-1"), vo.NewID("r-1"), "Menu")
	price, _ := vo.NewMoney(1000, "BRL")
	item := NewMenuItem(vo.NewID("item-1"), "Burger", "Juicy", price, "Main")

	m.AddItem(item)

	if len(m.Items()) != 1 {
		t.Errorf("expected 1 item, got %d", len(m.Items()))
	}
}

func TestMenu_ChangeItemAvailability(t *testing.T) {
	m := NewMenu(vo.NewID("m-1"), vo.NewID("r-1"), "Menu")
	itemID := vo.NewID("item-1")
	item := NewMenuItem(itemID, "Burger", "Juicy", vo.Money{}, "Main")
	m.AddItem(item)
	m.ClearEvents()

	m.ChangeItemAvailability(itemID, false)

	if item.IsAvailable() {
		t.Errorf("expected item to be unavailable")
	}
	if len(m.Events()) != 1 {
		t.Errorf("expected 1 event, got %d", len(m.Events()))
	}
	if m.Events()[0].EventType() != "restaurant.menu.item_availability_changed" {
		t.Errorf("expected event type restaurant.menu.item_availability_changed, got %s", m.Events()[0].EventType())
	}
}
