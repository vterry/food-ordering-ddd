package customer

import (
	"github.com/vterry/food-project/common/pkg/domain/vo"
	"testing"
)

func TestAddress_Getters(t *testing.T) {
	id := vo.NewID("addr1")
	custID := vo.NewID("cust1")
	street := "Main St"
	city := "Springfield"
	zip := "12345"
	isDefault := true

	addr := NewAddress(id, custID, street, city, zip, isDefault)

	if !addr.ID().Equals(id) {
		t.Errorf("expected ID %v, got %v", id, addr.ID())
	}
	if !addr.CustomerID().Equals(custID) {
		t.Errorf("expected CustomerID %v, got %v", custID, addr.CustomerID())
	}
	if addr.Street() != street {
		t.Errorf("expected street %s, got %s", street, addr.Street())
	}
	if addr.City() != city {
		t.Errorf("expected city %s, got %s", city, addr.City())
	}
	if addr.ZipCode() != zip {
		t.Errorf("expected zip %s, got %s", zip, addr.ZipCode())
	}
	if addr.IsDefault() != isDefault {
		t.Errorf("expected isDefault %v, got %v", isDefault, addr.IsDefault())
	}
}
