CREATE TABLE IF NOT EXISTS outbox_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    aggregate_id VARCHAR(36) NOT NULL,
    aggregate_type VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    payload JSON NOT NULL,
    occurred_on TIMESTAMP NOT NULL,
    published_at TIMESTAMP NULL,
    
    INDEX idx_outbox_unpublished (published_at)
);
