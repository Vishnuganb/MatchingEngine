-- name: CreateTrade :exec
INSERT INTO trade_capture_reports (
    trade_report_id,
    msg_type,
    exec_id,
    symbol,
    last_qty,
    last_px,
    trade_date,
    transact_time
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: CreateTradeSide :exec
INSERT INTO trade_sides (
    trade_report_id,
    side,
    order_id
)
VALUES ($1, $2, $3);

-- name: GetTrade :one
SELECT *
FROM trade_capture_reports
WHERE trade_report_id = $1;

-- name: GetTradeSides :many
SELECT *
FROM trade_sides
WHERE trade_report_id = $1;

-- name: ListTrades :many
SELECT *
FROM trade_capture_reports
ORDER BY trade_report_id;

-- name: ListTradeWithSides :many
SELECT t.*, s.id AS side_id, s.side, s.order_id
FROM trade_capture_reports t
         JOIN trade_sides s ON t.trade_report_id = s.trade_report_id
ORDER BY t.trade_report_id;

-- name: UpdateTrade :one
UPDATE trade_capture_reports
SET
    msg_type        = COALESCE($2, msg_type),
    exec_id         = COALESCE($3, exec_id),
    symbol          = COALESCE($4, symbol),
    last_qty        = COALESCE($5, last_qty),
    last_px         = COALESCE($6, last_px),
    trade_date      = COALESCE($7, trade_date),
    transact_time   = COALESCE($8, transact_time)
WHERE trade_report_id = $1
    RETURNING *;

-- name: UpdateTradeSide :one
UPDATE trade_sides
SET
    side     = COALESCE($2, side),
    order_id = COALESCE($3, order_id)
WHERE id = $1
    RETURNING *;

-- name: DeleteTrade :one
DELETE FROM trade_capture_reports
WHERE trade_report_id = $1
    RETURNING *;

-- name: DeleteTradeSidesByTradeID :many
DELETE FROM trade_sides
WHERE trade_report_id = $1
    RETURNING *;
