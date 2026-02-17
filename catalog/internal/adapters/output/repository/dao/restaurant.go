package dao

import "time"

type RestaurantDAO struct {
	ID   int64  `db:"id"`
	UUID string `db:"uuid"`
	Name string `db:"name"`

	AddressStreet  string `db:"address_street"`
	AddressNumber  string `db:"address_number"`
	AddressCompl   string `db:"address_compl"`
	AddressNeigh   string `db:"address_neigh"`
	AddressCity    string `db:"address_city"`
	AddressState   string `db:"address_state"`
	AddressZipCode string `db:"address_zipcode"`

	Status       string  `db:"status"`
	ActiveMenuID *string `db:"active_menu_id"`

	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
