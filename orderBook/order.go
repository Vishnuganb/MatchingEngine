package orderBook

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"time"
)

type EventType string

const (
	EventTypePendingNew  EventType = "pending_new"
	EventTypeNew         EventType = "new"
	EventTypeFill        EventType = "fill"
	EventTypePartialFill EventType = "partial_fill"
	EventTypeCanceled    EventType = "canceled"
	EventTypeRejected    EventType = "rejected"
)

type Order struct {
	ID            string          `json:"id"`
	Instrument    string          `json:"instrument"`
	Price         decimal.Decimal `json:"price"`
	OrderQty      decimal.Decimal `json:"order_qty"`
	LeavesQty     decimal.Decimal `json:"leaves_qty"`
	Timestamp     int64           `json:"timestamp"`
	IsBid         bool            `json:"is_bid"`
	OrderStatus   EventType       `json:"order_status"`
	ExecType      EventType       `json:"exec_type"`
	ExecQty       decimal.Decimal `json:"exec_qty"`
	KafkaProducer EventNotifier
}

type Orders []Order

func (o *Order) Side() Side {
	if o.IsBid {
		return Buy
	}
	return Sell
}

func newBaseOrder(t EventType, orderID string, price decimal.Decimal, isBid bool) Order {
	o := Order{
		ID:        orderID,
		Timestamp: time.Now().UnixNano(),
		ExecType:  t,
		Price:     price,
		IsBid:     isBid,
	}
	return o
}

func newBaseOrderEvent(t EventType, order *Order) Order {
	o := newBaseOrder(t, order.ID, order.Price, order.IsBid)
	o.OrderQty = order.OrderQty
	o.LeavesQty = order.LeavesQty
	if o.Price.IsPositive() {
		o.Price = order.Price
	}
	o.ExecQty = order.ExecQty
	o.Instrument = order.Instrument
	return o
}

func newOrderEvent(order *Order) Order {
	o := newBaseOrderEvent(EventTypeNew, order)
	o.ExecQty = decimal.Zero
	return o
}

func newFillOrderEvent(order *Order, qty, tradePrice decimal.Decimal) Order {
	typ := EventTypeFill
	if order.LeavesQty.IsPositive() {
		typ = EventTypePartialFill
	}
	o := newBaseOrderEvent(typ, order)
	o.ExecQty = qty
	if tradePrice.GreaterThan(o.Price) {
		o.Price = tradePrice
	}
	return o
}

func newCanceledOrderEvent(order *Order) Order {
	o := newBaseOrderEvent(EventTypeCanceled, order)
	o.LeavesQty = decimal.Zero
	return o
}

func newRejectedOrderEvent(or *OrderRequest) Order {
	o := newBaseOrder(EventTypeRejected, or.ID, or.Price, or.Side== Buy)
	o.OrderQty = or.Qty
	o.LeavesQty = decimal.Zero
	return o
}

func (o *Order) publishExecutionReport(order Order) {
	// Handle Event object
	if o.KafkaProducer != nil {
		eventJSON, _ := json.Marshal(order)
		err := o.KafkaProducer.NotifyEventAndTrade(order.ID, eventJSON)
		if err != nil {
			return
		}
	}
}
