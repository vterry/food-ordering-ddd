CREATE TABLE IF NOT EXISTS categories (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    menu_id BIGINT NOT NULL, -- Strong relationship (Entity inside Aggregate), using surrogate FK for performance
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_categories_menus FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE
);
