-- name: CreateTrade :one
INSERT INTO trades (id, buyer_order_id, seller_order_id, qty, price, instrument)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;

-- name: GetTrade :one
SELECT *
FROM trades
WHERE id = $1;

-- name: ListTrades :many
SELECT *
FROM trades
ORDER BY id;

-- name: UpdateTrade :one
UPDATE trades
SET buyer_order_id  = COALESCE($2, buyer_order_id),
    seller_order_id = COALESCE($3, seller_order_id),
    qty             = COALESCE($4, qty),
    price           = COALESCE($5, price),
    instrument      = COALESCE($6, instrument)
WHERE id = $1 RETURNING *;

-- name: DeleteTrade :one
DELETE
FROM trades
WHERE id = $1 RETURNING *;