-- name: InsertCustomer :exec
INSERT INTO customers (id, name, email, phone)
VALUES (?, ?, ?, ?);

-- name: UpdateCustomer :exec
UPDATE customers
SET name = ?, email = ?, phone = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: GetCustomerByID :one
SELECT id, name, email, phone, created_at, updated_at
FROM customers
WHERE id = ?;

-- name: GetCustomerByEmail :one
SELECT id, name, email, phone, created_at, updated_at
FROM customers
WHERE email = ?;

-- name: InsertAddress :exec
INSERT INTO addresses (id, customer_id, street, city, zip_code, is_default)
VALUES (?, ?, ?, ?, ?, ?);

-- name: ListAddressesByCustomerID :many
SELECT id, customer_id, street, city, zip_code, is_default
FROM addresses
WHERE customer_id = ?;

-- name: UpdateDefaultAddress :exec
UPDATE addresses
SET is_default = (id = ?), updated_at = CURRENT_TIMESTAMP
WHERE customer_id = ?;

-- name: ClearAddresses :exec
DELETE FROM addresses WHERE customer_id = ?;
