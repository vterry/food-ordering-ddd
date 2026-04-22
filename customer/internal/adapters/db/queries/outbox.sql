-- name: InsertOutboxMessage :exec
INSERT INTO outbox_messages (
    id, aggregate_type, aggregate_id, event_type, payload, correlation_id
) VALUES (
    ?, ?, ?, ?, ?, ?
);

-- name: ListUnpublishedMessages :many
SELECT * FROM outbox_messages
WHERE published_at IS NULL
ORDER BY created_at ASC
LIMIT ?;

-- name: MarkMessageAsPublished :exec
UPDATE outbox_messages
SET published_at = CURRENT_TIMESTAMP
WHERE id = ?;
