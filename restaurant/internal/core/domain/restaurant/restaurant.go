package restaurant

import (
	"time"

	"github.com/vterry/food-project/common/pkg/domain/base"
	"github.com/vterry/food-project/common/pkg/domain/vo"
)

type Address struct {
	Street  string
	City    string
	ZipCode string
}

type OperatingPeriod struct {
	DayOfWeek time.Weekday
	Open      string
	Close     string
}

type Restaurant struct {
	base.BaseAggregateRoot
	name           string
	address        Address
	operatingHours []OperatingPeriod
}

func NewRestaurant(id vo.ID, name string, address Address, hours []OperatingPeriod) *Restaurant {
	r := &Restaurant{
		name:           name,
		address:        address,
		operatingHours: hours,
	}
	r.SetID(id)
	return r
}

func (r *Restaurant) Name() string {
	return r.name
}

func (r *Restaurant) Address() Address {
	return r.address
}

func (r *Restaurant) OperatingHours() []OperatingPeriod {
	return r.operatingHours
}
