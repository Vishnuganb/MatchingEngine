// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: trades.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createTrade = `-- name: CreateTrade :exec
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
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

type CreateTradeParams struct {
	TradeReportID string         `json:"trade_report_id"`
	MsgType       string         `json:"msg_type"`
	ExecID        string         `json:"exec_id"`
	Symbol        string         `json:"symbol"`
	LastQty       pgtype.Numeric `json:"last_qty"`
	LastPx        pgtype.Numeric `json:"last_px"`
	TradeDate     string         `json:"trade_date"`
	TransactTime  int64          `json:"transact_time"`
}

func (q *Queries) CreateTrade(ctx context.Context, arg CreateTradeParams) error {
	_, err := q.db.Exec(ctx, createTrade,
		arg.TradeReportID,
		arg.MsgType,
		arg.ExecID,
		arg.Symbol,
		arg.LastQty,
		arg.LastPx,
		arg.TradeDate,
		arg.TransactTime,
	)
	return err
}

const createTradeSide = `-- name: CreateTradeSide :exec
INSERT INTO trade_sides (
    trade_report_id,
    side,
    order_id
)
VALUES ($1, $2, $3)
`

type CreateTradeSideParams struct {
	TradeReportID string `json:"trade_report_id"`
	Side          int16  `json:"side"`
	OrderID       string `json:"order_id"`
}

func (q *Queries) CreateTradeSide(ctx context.Context, arg CreateTradeSideParams) error {
	_, err := q.db.Exec(ctx, createTradeSide, arg.TradeReportID, arg.Side, arg.OrderID)
	return err
}

const deleteTrade = `-- name: DeleteTrade :one
DELETE FROM trade_capture_reports
WHERE trade_report_id = $1
    RETURNING trade_report_id, msg_type, exec_id, symbol, last_qty, last_px, trade_date, transact_time
`

func (q *Queries) DeleteTrade(ctx context.Context, tradeReportID string) (TradeCaptureReport, error) {
	row := q.db.QueryRow(ctx, deleteTrade, tradeReportID)
	var i TradeCaptureReport
	err := row.Scan(
		&i.TradeReportID,
		&i.MsgType,
		&i.ExecID,
		&i.Symbol,
		&i.LastQty,
		&i.LastPx,
		&i.TradeDate,
		&i.TransactTime,
	)
	return i, err
}

const deleteTradeSidesByTradeID = `-- name: DeleteTradeSidesByTradeID :many
DELETE FROM trade_sides
WHERE trade_report_id = $1
    RETURNING id, trade_report_id, side, order_id
`

func (q *Queries) DeleteTradeSidesByTradeID(ctx context.Context, tradeReportID string) ([]TradeSide, error) {
	rows, err := q.db.Query(ctx, deleteTradeSidesByTradeID, tradeReportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TradeSide{}
	for rows.Next() {
		var i TradeSide
		if err := rows.Scan(
			&i.ID,
			&i.TradeReportID,
			&i.Side,
			&i.OrderID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTrade = `-- name: GetTrade :one
SELECT trade_report_id, msg_type, exec_id, symbol, last_qty, last_px, trade_date, transact_time
FROM trade_capture_reports
WHERE trade_report_id = $1
`

func (q *Queries) GetTrade(ctx context.Context, tradeReportID string) (TradeCaptureReport, error) {
	row := q.db.QueryRow(ctx, getTrade, tradeReportID)
	var i TradeCaptureReport
	err := row.Scan(
		&i.TradeReportID,
		&i.MsgType,
		&i.ExecID,
		&i.Symbol,
		&i.LastQty,
		&i.LastPx,
		&i.TradeDate,
		&i.TransactTime,
	)
	return i, err
}

const getTradeSides = `-- name: GetTradeSides :many
SELECT id, trade_report_id, side, order_id
FROM trade_sides
WHERE trade_report_id = $1
`

func (q *Queries) GetTradeSides(ctx context.Context, tradeReportID string) ([]TradeSide, error) {
	rows, err := q.db.Query(ctx, getTradeSides, tradeReportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TradeSide{}
	for rows.Next() {
		var i TradeSide
		if err := rows.Scan(
			&i.ID,
			&i.TradeReportID,
			&i.Side,
			&i.OrderID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listTradeWithSides = `-- name: ListTradeWithSides :many
SELECT t.trade_report_id, t.msg_type, t.exec_id, t.symbol, t.last_qty, t.last_px, t.trade_date, t.transact_time, s.id AS side_id, s.side, s.order_id
FROM trade_capture_reports t
         JOIN trade_sides s ON t.trade_report_id = s.trade_report_id
ORDER BY t.trade_report_id
`

type ListTradeWithSidesRow struct {
	TradeReportID string         `json:"trade_report_id"`
	MsgType       string         `json:"msg_type"`
	ExecID        string         `json:"exec_id"`
	Symbol        string         `json:"symbol"`
	LastQty       pgtype.Numeric `json:"last_qty"`
	LastPx        pgtype.Numeric `json:"last_px"`
	TradeDate     string         `json:"trade_date"`
	TransactTime  int64          `json:"transact_time"`
	SideID        int32          `json:"side_id"`
	Side          int16          `json:"side"`
	OrderID       string         `json:"order_id"`
}

func (q *Queries) ListTradeWithSides(ctx context.Context) ([]ListTradeWithSidesRow, error) {
	rows, err := q.db.Query(ctx, listTradeWithSides)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListTradeWithSidesRow{}
	for rows.Next() {
		var i ListTradeWithSidesRow
		if err := rows.Scan(
			&i.TradeReportID,
			&i.MsgType,
			&i.ExecID,
			&i.Symbol,
			&i.LastQty,
			&i.LastPx,
			&i.TradeDate,
			&i.TransactTime,
			&i.SideID,
			&i.Side,
			&i.OrderID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listTrades = `-- name: ListTrades :many
SELECT trade_report_id, msg_type, exec_id, symbol, last_qty, last_px, trade_date, transact_time
FROM trade_capture_reports
ORDER BY trade_report_id
`

func (q *Queries) ListTrades(ctx context.Context) ([]TradeCaptureReport, error) {
	rows, err := q.db.Query(ctx, listTrades)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []TradeCaptureReport{}
	for rows.Next() {
		var i TradeCaptureReport
		if err := rows.Scan(
			&i.TradeReportID,
			&i.MsgType,
			&i.ExecID,
			&i.Symbol,
			&i.LastQty,
			&i.LastPx,
			&i.TradeDate,
			&i.TransactTime,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateTrade = `-- name: UpdateTrade :one
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
    RETURNING trade_report_id, msg_type, exec_id, symbol, last_qty, last_px, trade_date, transact_time
`

type UpdateTradeParams struct {
	TradeReportID string         `json:"trade_report_id"`
	MsgType       string         `json:"msg_type"`
	ExecID        string         `json:"exec_id"`
	Symbol        string         `json:"symbol"`
	LastQty       pgtype.Numeric `json:"last_qty"`
	LastPx        pgtype.Numeric `json:"last_px"`
	TradeDate     string         `json:"trade_date"`
	TransactTime  int64          `json:"transact_time"`
}

func (q *Queries) UpdateTrade(ctx context.Context, arg UpdateTradeParams) (TradeCaptureReport, error) {
	row := q.db.QueryRow(ctx, updateTrade,
		arg.TradeReportID,
		arg.MsgType,
		arg.ExecID,
		arg.Symbol,
		arg.LastQty,
		arg.LastPx,
		arg.TradeDate,
		arg.TransactTime,
	)
	var i TradeCaptureReport
	err := row.Scan(
		&i.TradeReportID,
		&i.MsgType,
		&i.ExecID,
		&i.Symbol,
		&i.LastQty,
		&i.LastPx,
		&i.TradeDate,
		&i.TransactTime,
	)
	return i, err
}

const updateTradeSide = `-- name: UpdateTradeSide :one
UPDATE trade_sides
SET
    side     = COALESCE($2, side),
    order_id = COALESCE($3, order_id)
WHERE id = $1
    RETURNING id, trade_report_id, side, order_id
`

type UpdateTradeSideParams struct {
	ID      int32  `json:"id"`
	Side    int16  `json:"side"`
	OrderID string `json:"order_id"`
}

func (q *Queries) UpdateTradeSide(ctx context.Context, arg UpdateTradeSideParams) (TradeSide, error) {
	row := q.db.QueryRow(ctx, updateTradeSide, arg.ID, arg.Side, arg.OrderID)
	var i TradeSide
	err := row.Scan(
		&i.ID,
		&i.TradeReportID,
		&i.Side,
		&i.OrderID,
	)
	return i, err
}
