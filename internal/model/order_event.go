package model

import "github.com/shopspring/decimal"

type Order struct {
	ID          string          `json:"id"`
	Instrument  string          `json:"instrument"`
	Price       decimal.Decimal `json:"price"`
	OrderQty    decimal.Decimal `json:"order_qty"`
	LeavesQty   decimal.Decimal `json:"leaves_qty"`
	Timestamp   int64           `json:"timestamp"`
	IsBid       bool            `json:"is_bid"`
	OrderStatus string          `json:"order_status"`
	ExecType    string          `json:"exec_type"`
	ExecQty     decimal.Decimal `json:"exec_qty"`
}

type Orders []Order
