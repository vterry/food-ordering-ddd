package enums

import (
	"fmt"
	"strings"
)

type MenuStatus int

const (
	MenuDraft MenuStatus = iota
	MenuActive
	MenuArchived
)

type ItemStatus int

const (
	ItemAvailable ItemStatus = iota
	ItemUnavailable
	ItemTempUnavailable
)

type RestaurantStatus int

const (
	RestaurantOpened RestaurantStatus = iota
	RestaurantClosed
)

func (s MenuStatus) String() string {
	return [...]string{"DRAFT", "ACTIVE", "ARCHIVED"}[s]
}

func (s MenuStatus) IsValid() bool {
	return s >= MenuDraft && s <= MenuArchived
}

func ParseMenuStatus(s string) (MenuStatus, error) {
	switch strings.ToUpper(s) {
	case "DRAFT":
		return MenuDraft, nil
	case "ACTIVE":
		return MenuActive, nil
	case "ARCHIVED":
		return MenuArchived, nil
	default:
		return MenuDraft, fmt.Errorf("invalid menu status: %s", s)
	}
}

func (s ItemStatus) String() string {
	return [...]string{"AVAILABLE", "UNAVAILABLE", "TEMP_UNAVAILABLE"}[s]
}

func (s ItemStatus) IsValid() bool {
	return s >= ItemAvailable && s <= ItemTempUnavailable
}

func ParseItemStatus(s string) (ItemStatus, error) {
	switch strings.ToUpper(s) {
	case "AVAILABLE":
		return ItemAvailable, nil
	case "UNAVAILABLE":
		return ItemUnavailable, nil
	case "TEMP_UNAVAILABLE":
		return ItemTempUnavailable, nil
	default:
		return ItemUnavailable, fmt.Errorf("invalid item status: %s", s)
	}
}

func (s RestaurantStatus) String() string {
	return [...]string{"OPEN", "CLOSE"}[s]
}
func (s RestaurantStatus) IsValid() bool {
	return s >= RestaurantOpened && s <= RestaurantClosed
}

func ParseRestaurantStatus(s string) (RestaurantStatus, error) {
	switch strings.ToUpper(s) {
	case "OPEN":
		return RestaurantOpened, nil
	case "CLOSE":
		return RestaurantClosed, nil
	default:
		return RestaurantClosed, fmt.Errorf("invalid restaurant status: %s", s)
	}
}
