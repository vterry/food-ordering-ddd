-- name: GetTicketByID :one
SELECT * FROM tickets WHERE id = ?;

-- name: InsertTicket :exec
INSERT INTO tickets (id, order_id, restaurant_id, status, reject_reason)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateTicketStatus :exec
UPDATE tickets SET status = ?, reject_reason = ? WHERE id = ?;

-- name: ListTicketItemsByTicketID :many
SELECT * FROM ticket_items WHERE ticket_id = ?;

-- name: InsertTicketItem :exec
INSERT INTO ticket_items (ticket_id, product_id, name, quantity)
VALUES (?, ?, ?, ?);
