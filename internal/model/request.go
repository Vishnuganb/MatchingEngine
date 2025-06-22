package model

import (
	"errors"

	"github.com/shopspring/decimal"
)

// MsgType FIX MsgType <35> - Message Type
type MsgType string

const (
	MsgTypeNew         MsgType = "D"  // New Order - Single
	MsgTypeCancel      MsgType = "F"  // Order Cancel Request
	MsgTypeExecRpt     MsgType = "8"  // Execution Report
	MsgTypeTradeReport MsgType = "AE" // Trade Capture Report
)

type OrderRequest struct {
	MsgType        MsgType            `json:"35"`
	NewOrderReq    NewOrderRequest    `json:"new_order,omitempty"`
	CancelOrderReq OrderCancelRequest `json:"cancel_order,omitempty"`
}

// BaseOrderRequest Common fields across different FIX messages
type BaseOrderRequest struct {
	MsgType      MsgType `json:"35"`
	ClOrdID      string  `json:"11"`           // FIX <11> - Unique client order ID
	Side         Side    `json:"54"`           // FIX <54> - 1=Buy, 2=Sell
	Symbol       string  `json:"55"`           // FIX <55> - Symbol
	TransactTime int64   `json:"60"`           // FIX <60> - Epoch ns
	Text         string  `json:"58,omitempty"` // FIX <58> - Optional free text
}

type NewOrderRequest struct {
	BaseOrderRequest
	OrderQty decimal.Decimal `json:"38"`           // FIX <38>
	Price    decimal.Decimal `json:"44,omitempty"` // FIX <44> - Required if Limit order
}

type OrderCancelRequest struct {
	BaseOrderRequest
	OrigClOrdID string `json:"41"` // FIX <41> - Original client order ID
}

func (or *NewOrderRequest) ValidateNewOrder() error {
	switch {
	case or.ClOrdID == "":
		return errors.New("missing client order ID")
	case !or.Side.IsValid():
		return errors.New("invalid order side")
	case or.Symbol == "":
		return errors.New("missing symbol")
	case or.OrderQty.IsZero() || or.OrderQty.IsNegative():
		return errors.New("invalid order quantity")
	case or.Price.IsNegative():
		return errors.New("invalid price")
	}
	return nil
}
