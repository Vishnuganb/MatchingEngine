package model

import (
	"github.com/shopspring/decimal"
)

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
	CumQty      decimal.Decimal `json:"cum_qty"`
}

type Orders []Order

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
