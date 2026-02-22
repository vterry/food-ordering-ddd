DROP TABLE IF EXISTS outbox_dlq;
ALTER TABLE outbox_events DROP COLUMN retry_count;