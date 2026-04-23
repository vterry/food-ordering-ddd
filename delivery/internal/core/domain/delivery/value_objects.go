package delivery

import (
	"errors"
)

var (
	ErrInvalidStatusTransition = errors.New("invalid delivery status transition")
	ErrCourierAlreadyAssigned  = errors.New("courier already assigned to this delivery")
	ErrDeliveryAlreadyPickedUp = errors.New("delivery already picked up")
	ErrDeliveryNotPickedUp     = errors.New("delivery cannot be completed/refused before pick up")
	ErrEmptyCourierID          = errors.New("courier ID cannot be empty")
	ErrEmptyCourierName        = errors.New("courier name cannot be empty")
)

type Status string

const (
	StatusScheduled Status = "SCHEDULED"
	StatusPickedUp  Status = "PICKED_UP"
	StatusDelivered Status = "DELIVERED"
	StatusCancelled Status = "CANCELLED"
	StatusRefused   Status = "REFUSED"
)

func (s Status) String() string {
	return string(s)
}

type Address struct {
	Street  string
	City    string
	ZipCode string
}

func NewAddress(street, city, zipCode string) Address {
	return Address{
		Street:  street,
		City:    city,
		ZipCode: zipCode,
	}
}

type CourierInfo struct {
	ID   string
	Name string
}

func NewCourierInfo(id, name string) (CourierInfo, error) {
	if id == "" {
		return CourierInfo{}, ErrEmptyCourierID
	}
	if name == "" {
		return CourierInfo{}, ErrEmptyCourierName
	}
	return CourierInfo{
		ID:   id,
		Name: name,
	}, nil
}
