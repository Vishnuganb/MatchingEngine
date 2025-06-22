-- name: CreateTrade :one
INSERT INTO trade_capture_reports (trade_report_id, exec_id, order_id, cl_ord_id, symbol, side, last_qty, last_px, trade_date, transact_time, previously_reported, text)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING *;

-- name: GetTrade :one
SELECT *
FROM trade_capture_reports
WHERE trade_report_id = $1;

-- name: ListTrades :many
SELECT *
FROM trade_capture_reports
ORDER BY trade_report_id;

-- name: UpdateTrade :one
UPDATE trade_capture_reports
SET exec_id             = COALESCE($2, exec_id),
    order_id            = COALESCE($3, order_id),
    cl_ord_id           = COALESCE($4, cl_ord_id),
    symbol              = COALESCE($5, symbol),
    side                = COALESCE($6, side),
    last_qty            = COALESCE($7, last_qty),
    last_px             = COALESCE($8, last_px),
    trade_date          = COALESCE($9, trade_date),
    transact_time       = COALESCE($10, transact_time),
    previously_reported = COALESCE($11, previously_reported),
    text                = COALESCE($12, text)
WHERE trade_report_id = $1 RETURNING *;

-- name: DeleteTrade :one
DELETE
FROM trade_capture_reports
WHERE trade_report_id = $1 RETURNING *;