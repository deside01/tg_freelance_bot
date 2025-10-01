-- name: CreateOrder :one
INSERT INTO orders (
    title, description, link, published_at, created_at, updated_at
)
VALUES (
    ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: ClearOrders :exec
DELETE FROM orders;