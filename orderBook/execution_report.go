package orderBook

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"time"
)

type ExecType string

const (
	ExecTypeNew      ExecType = "0"
	ExecTypeFill     ExecType = "2"
	ExecTypeCanceled ExecType = "4"
	ExecTypeRejected ExecType = "8"
)

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

func (e *ExecutionReport) ResetQuantities() {
	e.CumQty = decimal.Zero
	e.LeavesQty = e.OrderQty
}

func (e *ExecutionReport) SetOrdStatus(status OrderStatus) {
	e.OrdStatus = status
}

func newBaseExecutionReport(order *Order, execType ExecType) ExecutionReport {
	var side Side
	if order.IsBid {
		side = Buy
	} else {
		side = Sell
	}
	return ExecutionReport{
		MsgType:      "8",
		ExecID:       uuid.NewString(),
		OrderID:      order.ID,
		ClOrdID:      "", // optionally populate from order.ClientOrderID
		ExecType:     execType,
		OrdStatus:    order.OrderStatus,
		Symbol:       order.Instrument,
		Side:         side,
		OrderQty:     order.OrderQty,
		LastShares:   decimal.Zero, // updated later if it's a fill
		LastPx:       order.Price,  // updated later if it's a fill
		LeavesQty:    order.LeavesQty,
		CumQty:       order.CumQty,
		AvgPx:        order.AvgPx,
		TransactTime: time.Now().UnixNano(),
	}
}
