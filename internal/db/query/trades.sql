-- name: CreateTrade :one
INSERT INTO trades (trade_report_id, exec_id, order_id, secondary_order_id, cl_ord_id, symbol, side, last_qty, last_px, trade_date, transact_time, previously_reported, text)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING *;

-- name: GetTrade :one
SELECT *
FROM trades
WHERE trade_report_id = $1;

-- name: ListTrades :many
SELECT *
FROM trades
ORDER BY trade_report_id;

-- name: UpdateTrade :one
UPDATE trades
SET exec_id             = COALESCE($2, exec_id),
    order_id            = COALESCE($3, order_id),
    secondary_order_id  = COALESCE($4, secondary_order_id),
    cl_ord_id           = COALESCE($5, cl_ord_id),
    symbol              = COALESCE($6, symbol),
    side                = COALESCE($7, side),
    last_qty            = COALESCE($8, last_qty),
    last_px             = COALESCE($9, last_px),
    trade_date          = COALESCE($10, trade_date),
    transact_time       = COALESCE($11, transact_time),
    previously_reported = COALESCE($12, previously_reported),
    text                = COALESCE($13, text)
WHERE trade_report_id = $1 RETURNING *;

-- name: DeleteTrade :one
DELETE
FROM trades
WHERE trade_report_id = $1 RETURNING *;