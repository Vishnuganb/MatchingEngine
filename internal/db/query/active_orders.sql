-- name: CreateActiveOrder :one
INSERT INTO active_orders (id, side, order_qty, leaves_qty, price, instrument, cum_qty, exec_type, order_status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING *;

-- name: GetActiveOrder :one
SELECT *
FROM active_orders
WHERE id = $1;

-- name: ListActiveOrders :many
SELECT *
FROM active_orders
ORDER BY id;

-- name: UpdateActiveOrder :one
UPDATE active_orders
SET exec_type         = COALESCE(sqlc.narg(exec_type), exec_type),
    leaves_qty   = COALESCE(sqlc.narg(leaves_qty), leaves_qty),
    cum_qty     = COALESCE(sqlc.narg(cum_qty), cum_qty),
    price        = COALESCE(sqlc.narg(price), price),
    order_status = COALESCE(sqlc.narg(order_status), order_status)
WHERE id = $1 RETURNING *;

-- name: DeleteActiveOrder :one
DELETE
FROM active_orders
WHERE id = $1 RETURNING *;
