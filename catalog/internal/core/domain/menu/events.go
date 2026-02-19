package menu

import (
	"github.com/vterry/food-ordering/catalog/internal/core/domain/enums"
	"github.com/vterry/food-ordering/catalog/internal/core/domain/valueobjects"
	common "github.com/vterry/food-ordering/common/pkg"
)

type ItemSnapshot struct {
	ItemID      string `json:"item_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
}
type CategorySnapshot struct {
	CategoryID string         `json:"category_id"`
	Name       string         `json:"name"`
	Items      []ItemSnapshot `json:"items"`
}

type MenuSnapshot struct {
	MenuID       string             `json:"menu_id"`
	RestaurantID string             `json:"restaurant_id"`
	Name         string             `json:"name"`
	Categories   []CategorySnapshot `json:"categories"`
}

type MenuCreated struct {
	common.BaseEvent
	Name         string `json:"name"`
	RestaurantID string `json:"restaurant_id"`
}

func NewMenuCreated(menu Menu) MenuCreated {
	return MenuCreated{
		BaseEvent:    common.NewBaseEvent("MenuCreated", menu.MenuID.String()),
		Name:         menu.Name(),
		RestaurantID: menu.RestaurantID().String(),
	}
}

type MenuActivated struct {
	common.BaseEvent
	Menu MenuSnapshot `json:"menu"`
}

func NewMenuActivated(menu Menu) MenuActivated {
	var categories []CategorySnapshot
	for _, cat := range menu.Categories() {
		var items []ItemSnapshot
		for _, item := range cat.Items() {
			items = append(items, ItemSnapshot{
				ItemID:      item.ItemID.String(),
				Name:        item.Name(),
				Description: item.Description(),
				Price:       item.BasePrice().String(),
			})
		}
		categories = append(categories, CategorySnapshot{
			CategoryID: cat.CategoryID.String(),
			Name:       cat.Name(),
			Items:      items,
		})
	}

	return MenuActivated{
		BaseEvent: common.NewBaseEvent("MenuActivated", menu.MenuID.String()),
		Menu: MenuSnapshot{
			MenuID:       menu.MenuID.String(),
			RestaurantID: menu.RestaurantID().String(),
			Name:         menu.Name(),
			Categories:   categories,
		},
	}
}

type MenuArchived struct {
	common.BaseEvent
	MenuID string `json:"menu_id"`
}

func NewMenuArchived(menuID valueobjects.MenuID) MenuArchived {
	return MenuArchived{
		BaseEvent: common.NewBaseEvent("MenuArchived", menuID.String()),
		MenuID:    menuID.String(),
	}
}

type ItemMenuCreated struct {
	common.BaseEvent
	Name       string `json:"name"`
	CategoryID string `json:"category_id"`
	Price      string `json:"price"`
}

func NewItemMenuCreated(categoryId valueobjects.CategoryID, item ItemMenu) ItemMenuCreated {
	return ItemMenuCreated{
		BaseEvent:  common.NewBaseEvent("ItemMenuCreated", item.ItemID.String()),
		Name:       item.Name(),
		CategoryID: categoryId.String(),
		Price:      item.BasePrice().String(),
	}
}

type ItemMenuRemoved struct {
	common.BaseEvent
	Name       string `json:"name"`
	CategoryID string `json:"category_id"`
	Price      string `json:"price"`
}

func NewItemMenuRemoved(categoryId valueobjects.CategoryID, item ItemMenu) ItemMenuRemoved {
	return ItemMenuRemoved{
		BaseEvent:  common.NewBaseEvent("ItemMenuRemoved", item.ItemID.String()),
		Name:       item.Name(),
		CategoryID: categoryId.String(),
		Price:      item.BasePrice().String(),
	}
}

type ItemMenuAvailabilityChanged struct {
	common.BaseEvent
	CategoryID string `json:"category_id"`
	OldStatus  string `json:"old_status"`
	NewStatus  string `json:"new_status"`
}

func NewItemMenuAvailabilityChanged(categoryId valueobjects.CategoryID, itemID valueobjects.ItemID, oldStatus, newStatus enums.ItemStatus) ItemMenuAvailabilityChanged {
	return ItemMenuAvailabilityChanged{
		BaseEvent:  common.NewBaseEvent("ItemMenuAvailabilityChanged", itemID.String()),
		CategoryID: categoryId.String(),
		OldStatus:  oldStatus.String(),
		NewStatus:  newStatus.String(),
	}
}

type ItemMenuPriceChanged struct {
	common.BaseEvent
	CategoryID string `json:"category_id"`
	OldPrice   string `json:"old_price"`
	NewPrice   string `json:"new_price"`
}

func NewItemMenuPriceChanged(categoryId valueobjects.CategoryID, itemID valueobjects.ItemID, oldPrice, newPrice common.Money) ItemMenuPriceChanged {
	return ItemMenuPriceChanged{
		BaseEvent:  common.NewBaseEvent("ItemMenuPriceChanged", itemID.String()),
		CategoryID: categoryId.String(),
		OldPrice:   oldPrice.String(),
		NewPrice:   newPrice.String(),
	}
}

type ItemMenuNameChanged struct {
	common.BaseEvent
	CategoryID string `json:"category_id"`
	OldName    string `json:"old_name"`
	NewName    string `json:"new_name"`
}

func NewItemMenuNameChanged(categoryId valueobjects.CategoryID, itemID valueobjects.ItemID, oldName, newName string) ItemMenuNameChanged {
	return ItemMenuNameChanged{
		BaseEvent:  common.NewBaseEvent("ItemMenuNameChanged", itemID.String()),
		CategoryID: categoryId.String(),
		OldName:    oldName,
		NewName:    newName,
	}
}

type MenuCategoryAdded struct {
	common.BaseEvent
	Category CategorySnapshot `json:"category"`
}

func NewMenuCategoryAdded(category Category) MenuCategoryAdded {
	var items []ItemSnapshot
	for _, item := range category.Items() {
		items = append(items, ItemSnapshot{
			ItemID:      item.ItemID.String(),
			Name:        item.Name(),
			Description: item.Description(),
			Price:       item.BasePrice().String(),
		})
	}

	return MenuCategoryAdded{
		BaseEvent: common.NewBaseEvent("MenuCategoryAdded", category.CategoryID.String()),
		Category: CategorySnapshot{
			CategoryID: category.CategoryID.String(),
			Name:       category.Name(),
			Items:      items,
		},
	}
}

type MenuCategoryRemoved struct {
	common.BaseEvent
	Category CategorySnapshot `json:"category"`
}

func NewMenuCategoryRemoved(category Category) MenuCategoryRemoved {
	var items []ItemSnapshot
	for _, item := range category.Items() {
		items = append(items, ItemSnapshot{
			ItemID:      item.ItemID.String(),
			Name:        item.Name(),
			Description: item.Description(),
			Price:       item.BasePrice().String(),
		})
	}

	return MenuCategoryRemoved{
		BaseEvent: common.NewBaseEvent("MenuCategoryRemoved", category.CategoryID.String()),
		Category: CategorySnapshot{
			CategoryID: category.CategoryID.String(),
			Name:       category.Name(),
			Items:      items,
		},
	}
}
