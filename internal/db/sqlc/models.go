// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Execution struct {
	MsgType      string         `json:"msg_type"`
	ExecID       string         `json:"exec_id"`
	OrderID      string         `json:"order_id"`
	ClOrdID      pgtype.Text    `json:"cl_ord_id"`
	ExecType     string         `json:"exec_type"`
	OrdStatus    string         `json:"ord_status"`
	Symbol       string         `json:"symbol"`
	Side         string         `json:"side"`
	OrderQty     pgtype.Numeric `json:"order_qty"`
	LastShares   pgtype.Numeric `json:"last_shares"`
	LastPx       pgtype.Numeric `json:"last_px"`
	LeavesQty    pgtype.Numeric `json:"leaves_qty"`
	CumQty       pgtype.Numeric `json:"cum_qty"`
	AvgPx        pgtype.Numeric `json:"avg_px"`
	TransactTime int64          `json:"transact_time"`
	Text         pgtype.Text    `json:"text"`
}

type TradeCaptureReport struct {
	TradeReportID string         `json:"trade_report_id"`
	MsgType       string         `json:"msg_type"`
	ExecID        string         `json:"exec_id"`
	Symbol        string         `json:"symbol"`
	LastQty       pgtype.Numeric `json:"last_qty"`
	LastPx        pgtype.Numeric `json:"last_px"`
	TradeDate     string         `json:"trade_date"`
	TransactTime  int64          `json:"transact_time"`
}

type TradeSide struct {
	ID            int32  `json:"id"`
	TradeReportID string `json:"trade_report_id"`
	Side          int16  `json:"side"`
	OrderID       string `json:"order_id"`
}
