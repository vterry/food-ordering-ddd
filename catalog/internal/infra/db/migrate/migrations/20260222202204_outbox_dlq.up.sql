ALTER TABLE outbox_events ADD COLUMN retry_count INT NOT NULL DEFAULT 0;

CREATE TABLE outbox_dlq (
    id VARCHAR(36) PRIMARY KEY,
    aggregate_id VARCHAR(36) NOT NULL,
    aggregate_type VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSON NOT NULL,
    occurred_on DATETIME NOT NULL,
    retry_count INT NOT NULL DEFAULT 0,
    error_reason TEXT NOT NULL,
    moved_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);