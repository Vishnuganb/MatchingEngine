package model

import "github.com/shopspring/decimal"

type ExecutionReport struct {
	MsgType      string          `json:"35"`           // always "8"
	ExecID       string          `json:"17"`           // ExecID
	OrderID      string          `json:"37"`           // OrderID
	ClOrdID      string          `json:"11,omitempty"` // ClOrdID
	ExecType     ExecType        `json:"150"`          // ExecType
	OrdStatus    OrderStatus     `json:"39"`           // OrdStatus
	Symbol       string          `json:"55"`           // Symbol
	Side         Side            `json:"54"`           // Side
	OrderQty     decimal.Decimal `json:"38"`           // OrderQty
	LastShares   decimal.Decimal `json:"32"`           // LastShares
	LastPx       decimal.Decimal `json:"31"`           // LastPx
	LeavesQty    decimal.Decimal `json:"151"`          // LeavesQty
	CumQty       decimal.Decimal `json:"14"`           // CumQty
	AvgPx        decimal.Decimal `json:"6"`            // AvgPx
	TransactTime int64           `json:"60"`           // TransactTime
	Text         string          `json:"58,omitempty"` // Text
}
