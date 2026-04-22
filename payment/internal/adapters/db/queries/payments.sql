-- name: CreatePayment :exec
INSERT INTO payments (id, order_id, amount, status, created_at, updated_at) 
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdatePaymentStatus :exec
UPDATE payments SET status = ?, updated_at = ? WHERE id = ?;

-- name: GetPaymentByID :one
SELECT id, order_id, amount, status, created_at, updated_at FROM payments WHERE id = ? LIMIT 1;

-- name: GetPaymentByOrderID :one
SELECT id, order_id, amount, status, created_at, updated_at FROM payments WHERE order_id = ? LIMIT 1;

-- name: InsertOutboxMessage :exec
INSERT INTO outbox_messages (id, aggregate_type, aggregate_id, event_type, payload, correlation_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?);
