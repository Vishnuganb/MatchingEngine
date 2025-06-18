-- name: CreateExecution :one
INSERT INTO executions (id, order_id, side, order_qty, leaves_qty, price, instrument, cum_qty, exec_type, order_status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;

-- name: GetExecution :one
SELECT *
FROM executions
WHERE id = $1;

-- name: ListExecutions :many
SELECT *
FROM executions
ORDER BY id;

-- name: UpdateExecution :one
UPDATE executions
SET exec_type         = COALESCE(sqlc.narg(exec_type), exec_type),
    leaves_qty   = COALESCE(sqlc.narg(leaves_qty), leaves_qty),
    cum_qty     = COALESCE(sqlc.narg(cum_qty), cum_qty),
    price        = COALESCE(sqlc.narg(price), price),
    order_status = COALESCE(sqlc.narg(order_status), order_status)
WHERE id = $1 RETURNING *;

-- name: DeleteExecution :one
DELETE
FROM executions
WHERE id = $1 RETURNING *;
