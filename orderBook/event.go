package orderBook

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ExecType string

const (
	ExecTypePendingNew ExecType = "A"
	ExecTypeNew        ExecType = "0"
	ExecTypeFill       ExecType = "2"
	ExecTypeCanceled   ExecType = "4"
	ExecTypeRejected   ExecType = "8"
)

type Event struct {
	ID          string          `json:"id"`
	OrderID     string          `json:"order_id,omitempty"`
	Instrument  string          `json:"instrument"`
	Price       decimal.Decimal `json:"price"`
	OrderQty    decimal.Decimal `json:"order_qty"`
	LeavesQty   decimal.Decimal `json:"leaves_qty"`
	Timestamp   int64           `json:"timestamp"`
	IsBid       bool            `json:"is_bid"`
	OrderStatus OrderStatus     `json:"order_status"`
	ExecType    ExecType        `json:"exec_type"`
	CumQty      decimal.Decimal `json:"cum_qty"`
}

func NewOrderEvent(order *Order) {
	e := newBaseEvent(order, ExecTypeNew)
	e.CumQty = decimal.Zero
	e.OrderStatus = OrderStatusNew
	order.OrderStatus = OrderStatusNew
	order.publishExecutionReport(e)
}

func newFillEvent(order *Order, price, qty decimal.Decimal, isBid bool) {
	typ := ExecTypeFill
	orderStatus := OrderStatusFill
	if order.LeavesQty.IsPositive() {
		orderStatus = OrderStatusPartialFill
	}
	e := newBaseEvent(order, typ)
	e.Price = price
	e.LeavesQty = order.LeavesQty
	e.OrderStatus = orderStatus
	order.CumQty = order.CumQty.Add(qty)
	e.CumQty = qty
	order.OrderStatus = orderStatus
	e.IsBid = isBid
	order.publishExecutionReport(e)
}

func NewFillOrderEvent(order, matchOrder *Order, qty decimal.Decimal) {
	newFillEvent(order, matchOrder.Price, qty, order.IsBid)
	newFillEvent(matchOrder, matchOrder.Price, qty, matchOrder.IsBid)
}

func NewCanceledOrderEvent(order *Order) {
	log.Printf("Creating canceled event for order: %s", order.ID)
	e := newBaseEvent(order, ExecTypeCanceled)
	e.LeavesQty = decimal.Zero
	e.OrderStatus = OrderStatusCanceled
	order.OrderStatus = OrderStatusCanceled
	order.publishExecutionReport(e)
}

func NewRejectedOrderEvent(order *Order) {
	log.Printf("Creating Rejected event for order: %s", order.ID)
	e := newBaseEvent(order, ExecTypeRejected)
	e.LeavesQty = decimal.Zero
	e.OrderStatus = OrderStatusRejected
	order.OrderStatus = OrderStatusRejected
	order.publishExecutionReport(e)
}

func newBaseEvent(order *Order, execType ExecType) Event {
	return Event{
		ID:         uuid.NewString(),
		OrderID:    order.ID,
		Instrument: order.Instrument,
		Price:      order.Price,
		OrderQty:   order.OrderQty,
		LeavesQty:  order.LeavesQty,
		Timestamp:  time.Now().UnixNano(),
		IsBid:      order.IsBid,
		ExecType:   execType,
		CumQty:     order.CumQty,
	}
}
