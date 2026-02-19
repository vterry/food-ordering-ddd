
CREATE TABLE IF NOT EXISTS restaurant(
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  uuid VARCHAR(36) NOT NULL UNIQUE,
  name VARCHAR(40) NOT NULL,
  
  address_street VARCHAR(255) NOT NULL,
  address_number VARCHAR(50) NOT NULL,
  address_compl VARCHAR(255) NOT NULL DEFAULT '',
  address_neigh VARCHAR(255) NOT NULL,
  address_city VARCHAR(255) NOT NULL,
  address_state VARCHAR(2) NOT NULL,
  address_zipcode VARCHAR(20) NOT NULL,
  
  
  status VARCHAR(20) NOT NULL DEFAULT 'CLOSED',
  active_menu_id VARCHAR(36) NULL,
  
  
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  INDEX idx_restaurant_status(status)
);