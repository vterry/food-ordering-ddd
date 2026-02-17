CREATE TABLE IF NOT EXISTS items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    category_id BIGINT NOT NULL, -- Strong relationship (Entity inside Aggregate)
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price_cents BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'UNAVAILABLE',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_items_categories FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);
