-- name: CreateOrder :exec
INSERT INTO orders (id, customer_id, restaurant_id, status, total_amount, currency, delivery_address, correlation_id, version, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetOrder :one
SELECT * FROM orders WHERE id = ?;

-- name: ListOrdersByCustomer :many
SELECT * FROM orders WHERE customer_id = ? ORDER BY created_at DESC;

-- name: UpdateOrderWithVersion :exec
UPDATE orders 
SET status = ?, version = version + 1, updated_at = ?
WHERE id = ? AND version = ?;

-- name: CreateOrderItem :exec
INSERT INTO order_items (id, order_id, menu_item_id, name, quantity, unit_price, notes)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: ListOrderItems :many
SELECT * FROM order_items WHERE order_id = ?;

-- name: SaveSagaState :exec
INSERT INTO saga_state (order_id, current_step, status, data_json, updated_at)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE 
    current_step = VALUES(current_step),
    status = VALUES(status),
    data_json = VALUES(data_json),
    updated_at = VALUES(updated_at);

-- name: GetSagaState :one
SELECT * FROM saga_state WHERE order_id = ?;

-- name: SaveOutboxMessage :exec
INSERT INTO outbox_messages (id, aggregate_type, aggregate_id, event_type, payload, correlation_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetUnpublishedOutboxMessages :many
SELECT * FROM outbox_messages 
WHERE published_at IS NULL 
ORDER BY created_at ASC 
LIMIT ?;

-- name: MarkOutboxMessageAsPublished :exec
UPDATE outbox_messages 
SET published_at = NOW() 
WHERE id = ?;
