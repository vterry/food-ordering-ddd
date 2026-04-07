ALTER TABLE outbox_events ADD COLUMN claimed_by VARCHAR(255) NULL;
ALTER TABLE outbox_events ADD COLUMN claimed_at TIMESTAMP NULL;

CREATE INDEX idx_outbox_claimable ON outbox_events (published_at, claimed_by, claimed_at);
