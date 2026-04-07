package enums

import (
	"fmt"
	"strings"
)

type OrderStatus int

const (
	Pending OrderStatus = iota
	Paid
	Confirmed
	InDelivery
	Delivered
	Cancelled
	Failed
)

func (s OrderStatus) String() string {
	return [...]string{
		"PENDING",
		"PAID",
		"CONFIRMED",
		"IN_DELIVERY",
		"DELIVERED",
		"CANCELLED",
		"FAILED",
	}[s]
}

func ParseOrderStatus(s string) (OrderStatus, error) {
	switch strings.ToUpper(s) {
	case "PENDING":
		return Pending, nil
	case "PAID":
		return Paid, nil
	case "CONFIRMED":
		return Confirmed, nil
	case "IN_DELIVERY":
		return InDelivery, nil
	case "DELIVERED":
		return Delivered, nil
	case "CANCELLED":
		return Cancelled, nil
	case "FAILED":
		return Failed, nil
	default:
		return Pending, fmt.Errorf("invalid menu status: %s", s)
	}
}
