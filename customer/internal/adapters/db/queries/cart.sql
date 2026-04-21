-- name: GetCartByCustomerID :one
SELECT customer_id, restaurant_id, updated_at
FROM carts
WHERE customer_id = ?;

-- name: ListCartItemsByCustomerID :many
SELECT id, customer_id, product_id, name, price_amount, price_currency, quantity, observation
FROM cart_items
WHERE customer_id = ?;

-- name: InsertCart :exec
INSERT INTO carts (customer_id, restaurant_id)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE restaurant_id = VALUES(restaurant_id);

-- name: InsertCartItem :exec
INSERT INTO cart_items (customer_id, product_id, name, price_amount, price_currency, quantity, observation)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: DeleteCartItems :exec
DELETE FROM cart_items WHERE customer_id = ?;

-- name: DeleteCart :exec
DELETE FROM carts WHERE customer_id = ?;
