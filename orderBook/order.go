package orderBook

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"time"
)

type EventType string

const (
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
	OrderQty      decimal.Decimal `json:"order_qty,omitempty"`
	LeavesQty     decimal.Decimal `json:"leaves_qty"`
	Timestamp     int64           `json:"timestamp"`
	IsBid         bool            `json:"is_bid"`
	OrderStatus   EventType       `json:"order_status,omitempty"`
	ExecType      EventType       `json:"exec_type,omitempty"`
	ExecQty       decimal.Decimal `json:"exec_qty,omitempty"`
	KafkaProducer EventNotifier
}

type Orders []Order

func (o *Order) Side() Side {
	if o.IsBid {
		return Buy
	}
	return Sell
}

func newBaseOrder(t EventType, orderID string, side Side) Order {
	o := Order{
		ID:        orderID,
		Timestamp: time.Now().UnixNano(),
		ExecType:  t,
	}
	return o
}

func newBaseOrderEvent(t EventType, order *Order) Order {
	o := newBaseOrder(t, order.ID, order.Side())
	o.OrderQty = order.OrderQty
	o.LeavesQty = order.LeavesQty
	if o.Price.IsPositive() {
		o.Price = order.Price
	}
	if o.Instrument != "" {
		o.Instrument = order.Instrument
	}
	return o
}

func newOrderEvent(o *Order) Order {
	return newBaseOrderEvent(EventTypeNew, o)
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
	o := newBaseOrder(EventTypeRejected, or.ID, or.Side)
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
