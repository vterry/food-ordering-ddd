package input

type CreateMenuRequest struct {
	Name string `json:"name" validate:"required"`
}

type AddCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

type AddItemRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	PriceCents  int64  `json:"price_cents" validate:"required,gt=0"`
}

type CreateRestaurantRequest struct {
	Name    string         `json:"name" validate:"required"`
	Address AddressRequest `json:"address" validate:"required"`
}

type UpdateItemRequest struct {
	PriceCents int64  `json:"price_cents" validate:"required,gt=0"`
	Status     string `json:"status" validate:"required"`
}

type AddressRequest struct {
	Street       string `json:"street" validate:"required"`
	Number       string `json:"number" validate:"required"`
	Complement   string `json:"complement"`
	Neighborhood string `json:"neighborhood" validate:"required"`
	City         string `json:"city" validate:"required"`
	State        string `json:"state" validate:"required"`
	ZipCode      string `json:"zip_code" validate:"required"`
}

type MenuResponse struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	RestaurantID string              `json:"restaurant_id"`
	Status       string              `json:"status"`
	Categories   []*CategoryResponse `json:"categories"`
}

type RestaurantResponse struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Address      AddressResponse `json:"address"`
	Status       string          `json:"status"`
	ActiveMenuID string          `json:"active_menu_id,omitempty"`
}

type AddressResponse struct {
	Street       string `json:"street"`
	Number       string `json:"number"`
	Complement   string `json:"complement,omitempty"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	ZipCode      string `json:"zip_code"`
}

type CategoryResponse struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Items []*ItemResponse `json:"items"`
}

type ItemResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceCents  int64  `json:"price_cents"`
	Status      string `json:"status"`
}

type ValidateOrderRequest struct {
	RestaurantID string   `json:"restaurant_id"`
	ItemIDs      []string `json:"item_ids"`
}
type ValidateOrderResponse struct {
	Valid            bool           `json:"valid"`
	ValidationErrors []string       `json:"validation_errors"`
	Items            []ItemSnapshot `json:"items"`
}
type ItemSnapshot struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
}
