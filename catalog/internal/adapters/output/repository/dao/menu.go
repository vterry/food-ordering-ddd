package dao

import "time"

type MenuDAO struct {
	ID           int64     `db:"id"`
	UUID         string    `db:"uuid"`
	RestaurantID string    `db:"restaurant_id"`
	Name         string    `db:"name"`
	Status       string    `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type CategoryDAO struct {
	ID          int64     `db:"id"`
	UUID        string    `db:"uuid"`
	MenuID      int64     `db:"menu_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type ItemDAO struct {
	ID          int64     `db:"id"`
	UUID        string    `db:"uuid"`
	CategoryID  int64     `db:"category_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	PriceCents  int64     `db:"price_cents"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type MenuCompositeDAO struct {
	MenuID       int64
	MenuUUID     string
	RestaurantID string
	MenuName     string
	MenuStatus   string

	CategoryID   *int64
	CategoryUUID *string
	CategoryName *string

	ItemID     *int64
	ItemUUID   *string
	ItemName   *string
	ItemDesc   *string
	ItemPrice  *int64
	ItemStatus *string
}
