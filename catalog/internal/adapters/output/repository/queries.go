package repository

var (
	QueryUpsertMenu = `INSERT INTO menus (uuid, restaurant_id, name, status, created_at, updated_at) 
	VALUES (?,?,?,?,NOW(),NOW())
	ON DUPLICATE KEY UPDATE
		name = VALUES(name),
		status = VALUES(status),
		updated_at = NOW();`

	QueryGetMenuIDByUUID = `SELECT id FROM menus WHERE uuid = ?`

	QueryDeleteCategoriesByMenuID = `DELETE FROM categories WHERE menu_id = ?`

	QueryInsertCategory = `INSERT INTO categories (uuid, menu_id, name, description, created_at, updated_at)
	VALUES (?,?,?,?,NOW(),NOW()) `

	QueryInsertItem = `INSERT INTO items (uuid, category_id, name, description, price_cents, status, created_at, updated_at)
	VALUES (?,?,?,?,?,?,NOW(),NOW())`

	QueryFindFullMenuByUUID = `
		SELECT 
			m.id as menu_db_id, m.uuid as menu_uuid, m.restaurant_id, m.name as menu_name, m.status as menu_status,
			c.id as cat_db_id, c.uuid as cat_uuid, c.name as cat_name,
			i.id as item_db_id, i.uuid as item_uuid, i.name as item_name, i.description as item_desc, i.price_cents, i.status as item_status
		FROM menus m
		LEFT JOIN categories c ON c.menu_id = m.id
		LEFT JOIN items i ON i.category_id = c.id
		WHERE m.uuid = ?
        ORDER BY c.id, i.id
	`

	QueryFindAllMenusByRestaurantID = `
    SELECT 
        m.id as menu_db_id, m.uuid as menu_uuid, m.restaurant_id, m.name as menu_name, m.status as menu_status,
        c.id as cat_db_id, c.uuid as cat_uuid, c.name as cat_name,
        i.id as item_db_id, i.uuid as item_uuid, i.name as item_name, i.description as item_desc, i.price_cents, i.status as item_status
    FROM menus m
    LEFT JOIN categories c ON c.menu_id = m.id
    LEFT JOIN items i ON i.category_id = c.id
    WHERE m.restaurant_id = ? 
    ORDER BY m.id, c.id, i.id ASC
		`

	QueryFindActiveMenusByRestaurantID = `
    SELECT 
        m.id as menu_db_id, m.uuid as menu_uuid, m.restaurant_id, m.name as menu_name, m.status as menu_status,
        c.id as cat_db_id, c.uuid as cat_uuid, c.name as cat_name,
        i.id as item_db_id, i.uuid as item_uuid, i.name as item_name, i.description as item_desc, i.price_cents, i.status as item_status
    FROM menus m
    LEFT JOIN categories c ON c.menu_id = m.id
    LEFT JOIN items i ON i.category_id = c.id
    WHERE m.restaurant_id = ? 
		AND m.status = 'ACTIVE'
    ORDER BY m.id, c.id, i.id ASC
		`

	QueryUpsertRestaurant = `
	INSERT INTO restaurants (uuid, name, address_street, address_number, address_compl, address_neigh, address_city, address_state, address_zipcode, status, active_menu_id, created_at, updated_at)
	VALUES (?,?,?,?,?,?,?,?,?,?,?,NOW(),NOW())
	ON DUPLICATE KEY UPDATE
		name = VALUES(name),
		address_street = VALUES(address_street),
		address_number = VALUES(address_number),
		address_compl = VALUES(address_compl),
		address_neigh = VALUES(address_neigh),
		address_city = VALUES(address_city),
		address_state = VALUES(address_state),
		address_zipcode = VALUES(address_zipcode),
		status = VALUES(status),
		active_menu_id = VALUES(active_menu_id),
		updated_at = NOW();`

	QueryFindRestaurantByUUID = `
	SELECT id, uuid, name, address_street, address_number, address_compl, address_neigh, address_city, address_state, address_zipcode, status, active_menu_id, created_at, updated_at
	FROM restaurants
	WHERE uuid = ?`

	QueryFindAllRestaurants = `
	SELECT id, uuid, name, address_street, address_number, address_compl, address_neigh, address_city, address_state, address_zipcode, status, active_menu_id, created_at, updated_at
	FROM restaurants`

	QueryFindOrderValidationData = `
SELECT 
    r.uuid       AS restaurant_uuid,
    r.status     AS restaurant_status,
    (r.active_menu_id IS NOT NULL) AS has_active_menu,
    i.uuid       AS item_uuid,
    i.name       AS item_name,
    i.price_cents,
    i.status     AS item_status
FROM restaurants r
LEFT JOIN menus m ON m.uuid = r.active_menu_id
LEFT JOIN categories c ON c.menu_id = m.id
LEFT JOIN items i ON i.category_id = c.id AND i.uuid IN (%s)
WHERE r.uuid = ?;
	`

	QueryCatalogFindActiveMenuRows = `
	SELECT 
			m.uuid, m.name, m.status,
			c.uuid, c.name,
			i.uuid, i.name, i.description, i.price_cents, i.status
	FROM menus m
	LEFT JOIN categories c ON c.menu_id = m.id
	LEFT JOIN items i ON i.category_id = c.id
	WHERE m.restaurant_id = ? 
		AND m.status = 'ACTIVE' 
		AND (i.status = 'AVAILABLE' OR i.status IS NULL)
	ORDER BY c.id, i.id ASC
`

	QueryInsertOutboxEvent = `
INSERT INTO outbox_events (uuid, aggregate_id, aggregate_type, type, payload, occurred_on, published_at)
VALUES (?,?,?,?,?,?,NULL)`

	QueryFindUnpublishedEvents = `
SELECT id, uuid, aggregate_id, aggregate_type, type, payload, occurred_on, retry_count
FROM outbox_events
WHERE published_at IS NULL
ORDER BY id ASC
LIMIT ?
FOR UPDATE SKIP LOCKED
`

	QueryMarkAsPublished = `
UPDATE outbox_events
SET published_at = NOW()
WHERE uuid = ?`

	QueryIncrementRetryCount = `
UPDATE outbox_events SET retry_count = retry_count + 1 WHERE uuid = ? 
`

	QueryInsertOutboxDLQ = `
INSERT INTO outbox_dlq (id, aggregate_id, aggregate_type, event_type, payload, occurred_on, retry_count, error_reason)
VALUES (?,?,?,?,?,?,?,?)
`

	QueryDeleteFromOutbox = `
DELETE FROM outbox_events WHERE uuid = ?
`
)
