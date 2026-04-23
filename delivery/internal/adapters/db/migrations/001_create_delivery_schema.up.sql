CREATE TABLE deliveries (
    id              VARCHAR(255)  PRIMARY KEY,
    order_id        VARCHAR(255)  NOT NULL,
    restaurant_id   VARCHAR(255)  NOT NULL,
    customer_id     VARCHAR(255)  NOT NULL,
    status          VARCHAR(50)   NOT NULL DEFAULT 'SCHEDULED',
    street          VARCHAR(255)  NOT NULL,
    city            VARCHAR(100)  NOT NULL,
    zip_code        VARCHAR(20)   NOT NULL,
    courier_id      VARCHAR(255)  NULL,
    courier_name    VARCHAR(200)  NULL,
    created_at      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_deliveries_order (order_id),
    INDEX idx_deliveries_status (status)
);

CREATE TABLE outbox_messages (
    id              CHAR(36)     PRIMARY KEY,
    aggregate_type  VARCHAR(100) NOT NULL,
    aggregate_id    VARCHAR(255) NOT NULL,
    event_type      VARCHAR(100) NOT NULL,
    payload         JSON         NOT NULL,
    correlation_id  VARCHAR(255) NOT NULL,
    created_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at    TIMESTAMP    NULL,
    INDEX idx_outbox_unpublished (published_at, created_at)
);

CREATE TABLE processed_messages (
    message_id   VARCHAR(255)  PRIMARY KEY,
    processed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
