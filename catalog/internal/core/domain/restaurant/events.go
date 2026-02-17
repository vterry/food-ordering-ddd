package restaurant

import (
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	common "github.com/vterry/food-ordering/common/pkg"
)

type RestaurantCreated struct {
	common.BaseEvent
	Name    string `json:"name"`
	Street  string `json:"street"`
	Number  string `json:"number"`
	ZipCode string `json:"zip_code"`
	City    string `json:"city"`
	State   string `json:"state"`
	Status  string `json:"status"`
}

func NewRestaurantCreated(restaurant Restaurant) RestaurantCreated {
	return RestaurantCreated{
		BaseEvent: common.NewBaseEvent("RestaurantCreated", restaurant.ID().String()),
		Name:      restaurant.Name(),
		Street:    restaurant.address.Street(),
		Number:    restaurant.address.Number(),
		ZipCode:   restaurant.address.ZipCode(),
		City:      restaurant.address.City(),
		State:     restaurant.address.State(),
		Status:    restaurant.Status().String(),
	}
}

type RestaurantOpened struct {
	common.BaseEvent
	RestaurantID string `json:"restaurant_id"`
}

func NewRestaurantOpened(restaurantID valueobjects.RestaurantID) RestaurantOpened {
	return RestaurantOpened{
		BaseEvent:    common.NewBaseEvent("RestaurantOpened", restaurantID.String()),
		RestaurantID: restaurantID.String(),
	}
}

type RestaurantClosed struct {
	common.BaseEvent
	RestaurantID string `json:"restaurant_id"`
}

func NewRestaurantClosed(restaurantID valueobjects.RestaurantID) RestaurantClosed {
	return RestaurantClosed{
		BaseEvent:    common.NewBaseEvent("RestaurantClosed", restaurantID.String()),
		RestaurantID: restaurantID.String(),
	}
}

type RestaurantMenuUpdated struct {
	common.BaseEvent
	MenuID string `json:"menu_id"`
}

func NewRestaurantMenuUpdated(restaurantID valueobjects.RestaurantID, menuID valueobjects.MenuID) RestaurantMenuUpdated {
	return RestaurantMenuUpdated{
		BaseEvent: common.NewBaseEvent("RestaurantMenuUpdated", restaurantID.String()),
		MenuID:    menuID.String(),
	}
}
