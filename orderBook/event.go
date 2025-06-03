package orderBook

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type EventType string

const (
	EventTypeNew         EventType = "new"
	EventTypeFill        EventType = "fill"
	EventTypePartialFill EventType = "partial_fill"
	EventTypeCanceled    EventType = "canceled"
	EventTypeRejected    EventType = "rejected"
)

type Event struct {
	ID         string          `json:"id,omitempty"`
	OrderID    string          `json:"order_id,omitempty"`
	Instrument string          `json:"instrument,omitempty"`
	Timestamp  int64           `json:"timestamp,omitempty"`
	Type       EventType       `json:"type,omitempty"`
	Side       Side            `json:"side,omitempty"`
	Price      decimal.Decimal `json:"price,omitempty"`
	OrderQty   decimal.Decimal `json:"order_qty,omitempty"`
	LeavesQty  decimal.Decimal `json:"leaves_qty,omitempty"`
	ExecQty    decimal.Decimal `json:"exec_qty,omitempty"`
}

func newBaseEvent(t EventType, orderID string, side Side) Event {
	e := Event{
		ID:        uuid.NewString(),
		OrderID:   orderID,
		Timestamp: time.Now().UnixNano(),
		Type:      t,
		Side:      side,
	}
	return e
}

func newBaseOrderEvent(t EventType, o *Order) Event {
	e := newBaseEvent(t, o.ID, o.Side())
	e.OrderQty = o.Qty
	e.LeavesQty = o.LeavesQty
	if o.Price.IsPositive() {
		e.Price = o.Price
	}
	return e
}

func newEvent(o *Order) Event {
	return newBaseOrderEvent(EventTypeNew, o)
}

func newFillEvent(o *Order, qty, tradePrice decimal.Decimal) Event {
	typ := EventTypeFill
	if o.LeavesQty.IsPositive() {
		typ = EventTypePartialFill
	}
	e := newBaseOrderEvent(typ, o)
	e.ExecQty = qty
	if tradePrice.GreaterThan(o.Price){
		e.Price = tradePrice
	}
	return e
}

func newCanceledEvent(o *Order) Event {
	e := newBaseOrderEvent(EventTypeCanceled, o)
	return e
}

func newRejectedEvent(or *OrderRequest) Event {
	e := newBaseEvent(EventTypeRejected, or.ID, or.Side)
	e.OrderQty = or.Qty
	return e
}
