-- Common Tables
CREATE TABLE IF NOT EXISTS outbox_messages (
    id CHAR(36) PRIMARY KEY,
    aggregate_type VARCHAR(100) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSON NOT NULL,
    correlation_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP NULL,
    INDEX idx_outbox_unpublished (published_at, created_at)
);

-- Customer Service Tables
CREATE TABLE IF NOT EXISTS customers (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE,
    phone VARCHAR(20) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS addresses (
    id VARCHAR(255) PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL,
    street VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    zip_code VARCHAR(20) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS carts (
    customer_id VARCHAR(255) PRIMARY KEY,
    restaurant_id VARCHAR(255),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS cart_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    customer_id VARCHAR(255) NOT NULL,
    product_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    price_amount DECIMAL(10,2) NOT NULL,
    price_currency VARCHAR(3) NOT NULL,
    quantity INT NOT NULL,
    observation TEXT,
    FOREIGN KEY (customer_id) REFERENCES carts(customer_id) ON DELETE CASCADE
);

-- Restaurant Service Tables
CREATE TABLE IF NOT EXISTS restaurants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    street VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    zip_code VARCHAR(20) NOT NULL,
    operating_hours JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS menus (
    id VARCHAR(255) PRIMARY KEY,
    restaurant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS menu_items (
    id VARCHAR(255) PRIMARY KEY,
    menu_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price_amount DECIMAL(10,2) NOT NULL,
    price_currency VARCHAR(3) NOT NULL,
    category VARCHAR(100) NOT NULL,
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tickets (
    id VARCHAR(255) PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL,
    restaurant_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    reject_reason VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (restaurant_id) REFERENCES restaurants(id)
);

CREATE TABLE IF NOT EXISTS ticket_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    ticket_id VARCHAR(255) NOT NULL,
    product_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    quantity INT NOT NULL,
    FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE CASCADE
);

-- Payment Service Tables
CREATE TABLE IF NOT EXISTS payments (
    id VARCHAR(255) PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_payments_order_id (order_id)
);

-- Ordering Service Tables
CREATE TABLE IF NOT EXISTS orders (
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

CREATE TABLE IF NOT EXISTS order_items (
    id            VARCHAR(255)  PRIMARY KEY,
    order_id      VARCHAR(255)  NOT NULL,
    menu_item_id  VARCHAR(255)  NOT NULL,
    name          VARCHAR(200)  NOT NULL,
    quantity      INT           NOT NULL,
    unit_price    BIGINT        NOT NULL,
    notes         TEXT,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS saga_state (
    order_id      VARCHAR(255) PRIMARY KEY,
    current_step  VARCHAR(50)  NOT NULL,
    status        VARCHAR(20)  NOT NULL DEFAULT 'RUNNING',
    data_json     JSON         NOT NULL,
    updated_at    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE
);

-- Delivery Service Tables
CREATE TABLE IF NOT EXISTS deliveries (
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

CREATE TABLE IF NOT EXISTS processed_messages (
    message_id   VARCHAR(255)  PRIMARY KEY,
    processed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Test Data
INSERT INTO restaurants (id, name, street, city, zip_code, operating_hours) 
VALUES ('res-1', 'Test Restaurant', 'Main St', 'Test City', '12345', '[]');

INSERT INTO menus (id, restaurant_id, name, is_active)
VALUES ('menu-1', 'res-1', 'Standard Menu', TRUE);

INSERT INTO menu_items (id, menu_id, name, description, price_amount, price_currency, category, is_available)
VALUES ('p1', 'menu-1', 'Burger', 'A delicious burger', 10.50, 'USD', 'Main', TRUE);

INSERT INTO menu_items (id, menu_id, name, description, price_amount, price_currency, category, is_available)
VALUES ('p-unhappy', 'menu-1', 'REJECT_ME', 'Trigger rejection', 20.00, 'USD', 'Main', TRUE);
