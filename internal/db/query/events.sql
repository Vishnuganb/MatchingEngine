-- name: CreateEvent :one
INSERT INTO events (
    order_id, type, side, order_qty, leaves_qty, exec_qty, price
) VALUES (
             $1, $2, $3, $4, $5, $6, $7
         )
    RETURNING *;

-- name: GetEvent :one
SELECT * FROM events
WHERE id = $1;

-- name: ListEvents :many
SELECT * FROM events
ORDER BY timestamp DESC;

-- name: UpdateEvent :one
UPDATE events
SET
    order_id = COALESCE(sqlc.narg(order_id), order_id),
    type = COALESCE(sqlc.narg(type), type),
    side = COALESCE(sqlc.narg(side), side),
    order_qty = COALESCE(sqlc.narg(order_qty), order_qty),
    leaves_qty = COALESCE(sqlc.narg(leaves_qty), leaves_qty),
    exec_qty = COALESCE(sqlc.narg(exec_qty), exec_qty),
    price = COALESCE(sqlc.narg(price), price)
WHERE id = $1
    RETURNING *;

-- name: DeleteEvent :one
DELETE FROM events
WHERE id = $1
    RETURNING *;
