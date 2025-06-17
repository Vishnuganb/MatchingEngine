package orderBook

import (
	"github.com/shopspring/decimal"
)

type ExecutionReport struct {
	OrderID    string          `json:"order_id"`
	ExecType   string          `json:"exec_type"`
	Price      decimal.Decimal `json:"price"`
	OrderQty   decimal.Decimal `json:"order_qty"`
	CumQty     decimal.Decimal `json:"exec_qty"`
	LeavesQty  decimal.Decimal `json:"leaves_qty"`
	Instrument string          `json:"instrument"`
	IsBid      bool            `json:"is_bid"`
	Timestamp  int64           `json:"timestamp"`
}
