-- name: CreateExecution :exec
INSERT INTO executions (exec_id, order_id, cl_ord_id, exec_type, ord_status, symbol, side, order_qty, last_shares, last_px, leaves_qty, cum_qty, avg_px, transact_time, text, msg_type)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16);

-- name: GetExecution :one
SELECT *
FROM executions
WHERE exec_id = $1;

-- name: ListExecutions :many
SELECT *
FROM executions
ORDER BY exec_id;

-- name: UpdateExecution :one
UPDATE executions
SET cl_ord_id     = COALESCE($2, cl_ord_id),
    exec_type     = COALESCE($3, exec_type),
    ord_status    = COALESCE($4, ord_status),
    symbol        = COALESCE($5, symbol),
    side          = COALESCE($6, side),
    order_qty     = COALESCE($7, order_qty),
    last_shares   = COALESCE($8, last_shares),
    last_px       = COALESCE($9, last_px),
    leaves_qty    = COALESCE($10, leaves_qty),
    cum_qty       = COALESCE($11, cum_qty),
    avg_px        = COALESCE($12, avg_px),
    transact_time = COALESCE($13, transact_time),
    text          = COALESCE($14, text)
WHERE exec_id = $1 RETURNING *;

-- name: DeleteExecution :one
DELETE
FROM executions
WHERE exec_id = $1 RETURNING *;