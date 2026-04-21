package customer

import (
	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

type CustomerRegisteredEvent struct {
	base.BaseDomainEvent
	CustomerID vo.ID
	Name       string
	Email      string
}

func NewCustomerRegisteredEvent(customerID vo.ID, name, email string) CustomerRegisteredEvent {
	return CustomerRegisteredEvent{
		BaseDomainEvent: base.NewBaseDomainEvent(vo.NewID("event-" + customerID.String())),
		CustomerID:      customerID,
		Name:            name,
		Email:           email,
	}
}

func (e CustomerRegisteredEvent) EventType() string {
	return "CustomerRegistered"
}
