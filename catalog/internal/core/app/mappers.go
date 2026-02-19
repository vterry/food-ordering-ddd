package app

import (
	"github.com/vterry/food-ordering/catalog/internal/core/domain/menu"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/restaurant"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	"github.com/vterry/food-ordering/catalog/internal/core/ports/input"
)

func toRestaurantResponse(restaurant *restaurant.Restaurant) *input.RestaurantResponse {
	addressResponse := parseAddress(restaurant.Address())

	return &input.RestaurantResponse{
		ID:           restaurant.RestaurantID.String(),
		Name:         restaurant.Name(),
		Address:      *addressResponse,
		Status:       restaurant.Status().String(),
		ActiveMenuID: restaurant.ActiveMenuID().String(),
	}
}

func toMenuResponse(menu *menu.Menu) *input.MenuResponse {
	categoriesResponse := make([]*input.CategoryResponse, 0)
	categories := menu.Categories()

	for i := range categories {
		category := parseCategories(&categories[i])
		categoriesResponse = append(categoriesResponse, category)
	}

	return &input.MenuResponse{
		ID:           menu.MenuID.String(),
		Name:         menu.Name(),
		RestaurantID: menu.RestaurantID().String(),
		Status:       menu.Status().String(),
		Categories:   categoriesResponse,
	}
}

func parseAddress(address valueobjects.Address) *input.AddressResponse {
	return &input.AddressResponse{
		Street:       address.Street(),
		Number:       address.Number(),
		Complement:   address.Complement(),
		Neighborhood: address.Neighborhood(),
		City:         address.City(),
		State:        address.State(),
		ZipCode:      address.ZipCode(),
	}
}

func parseCategories(category *menu.Category) *input.CategoryResponse {

	itemsResponse := make([]*input.ItemResponse, 0)
	items := category.Items()
	for i := range items {
		item := parseItemResponse(items[i])
		itemsResponse = append(itemsResponse, item)
	}

	return &input.CategoryResponse{
		ID:    category.CategoryID.String(),
		Name:  category.Name(),
		Items: itemsResponse,
	}
}

func parseItemResponse(item menu.ItemMenu) *input.ItemResponse {
	return &input.ItemResponse{
		ID:          item.ItemID.String(),
		Name:        item.Name(),
		Description: item.Description(),
		PriceCents:  item.BasePrice().Amount(),
		Status:      item.Status().String(),
	}
}
