-- name: InsertOutboxMessage :exec
INSERT INTO outbox_messages (id, aggregate_type, aggregate_id, event_type, payload, correlation_id)
VALUES (?, ?, ?, ?, ?, ?);
