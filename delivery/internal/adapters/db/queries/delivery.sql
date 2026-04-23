-- name: SaveDelivery :exec
INSERT INTO deliveries (
    id, order_id, restaurant_id, customer_id, status, street, city, zip_code, courier_id, courier_name
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) ON DUPLICATE KEY UPDATE
    status = VALUES(status),
    courier_id = VALUES(courier_id),
    courier_name = VALUES(courier_name),
    updated_at = CURRENT_TIMESTAMP;

-- name: GetDeliveryByID :one
SELECT * FROM deliveries WHERE id = ? LIMIT 1;

-- name: InsertOutboxMessage :exec
INSERT INTO outbox_messages (
    id, aggregate_type, aggregate_id, event_type, payload, correlation_id
) VALUES (
    ?, ?, ?, ?, ?, ?
);

-- name: GetUnpublishedOutboxMessages :many
SELECT * FROM outbox_messages WHERE published_at IS NULL ORDER BY created_at ASC;

-- name: MarkOutboxMessageAsPublished :exec
UPDATE outbox_messages SET published_at = CURRENT_TIMESTAMP WHERE id = ?;

-- name: IsMessageProcessed :one
SELECT EXISTS(SELECT 1 FROM processed_messages WHERE message_id = ?);

-- name: MarkMessageAsProcessed :exec
INSERT INTO processed_messages (message_id) VALUES (?);
