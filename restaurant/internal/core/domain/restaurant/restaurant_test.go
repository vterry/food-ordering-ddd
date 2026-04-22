package restaurant

import (
	"testing"
	"time"

	"github.com/vterry/food-project/common/pkg/domain/vo"
)

func TestNewRestaurant(t *testing.T) {
	id := vo.NewID("rest-1")
	name := "Taste of Go"
	address := Address{
		Street:  "Gopher Way 123",
		City:    "Cloud City",
		ZipCode: "12345",
	}
	hours := []OperatingPeriod{
		{
			DayOfWeek: time.Monday,
			Open:      "08:00",
			Close:     "22:00",
		},
	}

	rest := NewRestaurant(id, name, address, hours)

	if !rest.ID().Equals(id) {
		t.Errorf("expected ID %v, got %v", id, rest.ID())
	}
	if rest.Name() != name {
		t.Errorf("expected name %s, got %s", name, rest.Name())
	}
	if rest.Address().Street != address.Street {
		t.Errorf("expected street %s, got %s", address.Street, rest.Address().Street)
	}
	if len(rest.OperatingHours()) != 1 {
		t.Errorf("expected 1 operating period, got %d", len(rest.OperatingHours()))
	}
}
