CREATE TABLE orders (
    id              VARCHAR(255)  PRIMARY KEY,
    customer_id     VARCHAR(255)  NOT NULL,
    restaurant_id   VARCHAR(255)  NOT NULL,
    status          VARCHAR(50)   NOT NULL DEFAULT 'CREATED',
    total_amount    BIGINT        NOT NULL,
    currency        VARCHAR(3)    NOT NULL DEFAULT 'BRL',
    delivery_address TEXT         NOT NULL,
    correlation_id  VARCHAR(255)  NOT NULL,
    version         INT           NOT NULL DEFAULT 1,
    created_at      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_orders_customer (customer_id),
    INDEX idx_orders_status (status)
);

CREATE TABLE order_items (
    id            VARCHAR(255)  PRIMARY KEY,
    order_id      VARCHAR(255)  NOT NULL,
    menu_item_id  VARCHAR(255)  NOT NULL,
    name          VARCHAR(200)  NOT NULL,
    quantity      INT           NOT NULL,
    unit_price    BIGINT        NOT NULL,
    notes         TEXT,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

CREATE TABLE saga_state (
    order_id      VARCHAR(255) PRIMARY KEY,
    current_step  VARCHAR(50)  NOT NULL,
    status        VARCHAR(20)  NOT NULL DEFAULT 'RUNNING',
    data_json     JSON         NOT NULL,
    updated_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
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
