-- name: GetRestaurantByID :one
SELECT * FROM restaurants WHERE id = ?;

-- name: InsertRestaurant :exec
INSERT INTO restaurants (id, name, street, city, zip_code, operating_hours)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateRestaurant :exec
UPDATE restaurants SET name = ?, street = ?, city = ?, zip_code = ?, operating_hours = ?
WHERE id = ?;
