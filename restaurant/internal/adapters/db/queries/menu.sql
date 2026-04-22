-- name: GetMenuByID :one
SELECT * FROM menus WHERE id = ?;

-- name: GetActiveMenuByRestaurantID :one
SELECT * FROM menus WHERE restaurant_id = ? AND is_active = TRUE;

-- name: InsertMenu :exec
INSERT INTO menus (id, restaurant_id, name, is_active)
VALUES (?, ?, ?, ?);

-- name: UpdateMenu :exec
UPDATE menus SET name = ?, is_active = ?
WHERE id = ?;

-- name: ListMenuItemsByMenuID :many
SELECT * FROM menu_items WHERE menu_id = ?;

-- name: InsertMenuItem :exec
INSERT INTO menu_items (id, menu_id, name, description, price_amount, price_currency, category, is_available)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateMenuItem :exec
UPDATE menu_items SET name = ?, description = ?, price_amount = ?, price_currency = ?, category = ?, is_available = ?
WHERE id = ?;

-- name: ClearMenuItems :exec
DELETE FROM menu_items WHERE menu_id = ?;
