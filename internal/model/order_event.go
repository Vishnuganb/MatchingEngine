package model

import (
	"github.com/shopspring/decimal"
)

type ExecutionReport struct {
	OrderID     string          `json:"order_id"`
	ExecType    string          `json:"exec_type"`
	Price       decimal.Decimal `json:"price"`
	OrderQty    decimal.Decimal `json:"order_qty"`
	CumQty      decimal.Decimal `json:"exec_qty"`
	LeavesQty   decimal.Decimal `json:"leaves_qty"`
	Instrument  string          `json:"instrument"`
	IsBid       bool            `json:"is_bid"`
	Timestamp   int64           `json:"timestamp"`
	OrderStatus string          `json:"order_status"`
}

type Trade struct {
	BuyerOrderID  string          `json:"buyer_order_id"`
	SellerOrderID string          `json:"seller_order_id"`
	Quantity      decimal.Decimal `json:"quantity"`
	Price         decimal.Decimal `json:"price"`
	Timestamp     int64           `json:"timestamp"`
	Instrument    string          `json:"instrument"`
}
