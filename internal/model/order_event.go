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
	CumQty     decimal.Decimal `json:"cum_qty"`
}

type Orders []Order

type OrderEvent struct {
	EventType   string          `json:"event_type"`
	OrderID     string          `json:"order_id"`
	Instrument  string          `json:"instrument"`
	Price       decimal.Decimal `json:"price"`
	Quantity    decimal.Decimal `json:"quantity"`
	LeavesQty   decimal.Decimal `json:"leaves_qty"`
	ExecQty     decimal.Decimal `json:"exec_qty"`
	IsBid       bool            `json:"is_bid"`
	OrderStatus string          `json:"order_status"`
	ExecType    string          `json:"exec_type"`
}

type Trade struct {
	BuyerOrderID  string `json:"buyer_order_id"`
	SellerOrderID string `json:"seller_order_id"`
	Quantity      uint64 `json:"quantity"`
	Price         uint64 `json:"price"`
	Timestamp     int64  `json:"timestamp"`
}
