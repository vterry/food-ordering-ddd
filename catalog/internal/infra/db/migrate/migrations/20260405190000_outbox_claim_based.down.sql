DROP INDEX idx_outbox_claimable ON outbox_events;
ALTER TABLE outbox_events DROP COLUMN claimed_at;
ALTER TABLE outbox_events DROP COLUMN claimed_by;
