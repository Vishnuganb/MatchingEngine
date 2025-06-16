package orderBook

import (
	"github.com/google/uuid"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

type ExecType string

const (
	ExecTypePendingNew  ExecType = "A"
	ExecTypeNew         ExecType = "0"
	ExecTypeFill        ExecType = "2"
	ExecTypePartialFill ExecType = "1"
	ExecTypeCanceled    ExecType = "4"
	ExecTypeRejected    ExecType = "8"
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
		typ = ExecTypePartialFill
		orderStatus = OrderStatusPartialFill
	}
	e := newBaseEvent(order, typ)
	e.Price = price
	e.LeavesQty = qty
	e.OrderStatus = orderStatus
	e.CumQty = order.CumQty
	order.OrderStatus = orderStatus
	order.publishExecutionReport(e)
}

func NewFillOrderEvent(order, matchOrder *Order, qty decimal.Decimal) {
	newFillEvent(order, order.Price, qty, order.IsBid)
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

func NewRejectedOrderEvent(req *OrderRequest) {
	_ = Event{
		ID:          uuid.NewString(),
		OrderID:     req.ID,
		Price:       req.Price,
		OrderQty:    req.Qty,
		LeavesQty:   decimal.Zero,
		Timestamp:   time.Now().UnixNano(),
		IsBid:       req.Side == Buy,
		OrderStatus: OrderStatusRejected,
		ExecType:    ExecTypeRejected,
		CumQty:      decimal.Zero,
	}
	//req.publishExecutionReport(e)
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
