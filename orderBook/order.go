package orderBook

import (
	"encoding/json"
	"log"
	"time"

	"github.com/shopspring/decimal"

	"MatchingEngine/internal/model"
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

func newBaseOrder(t EventType, orderID string, price decimal.Decimal, isBid bool, producer EventNotifier) Order {
	o := Order{
		ID:            orderID,
		Timestamp:     time.Now().UnixNano(),
		ExecType:      t,
		Price:         price,
		IsBid:         isBid,
		KafkaProducer: producer,
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
	o.OrderStatus = EventTypeNew
	o.publishEvent(EventTypeNew)
	return o
}

func newFillOrderEvent(order *Order, qty, tradePrice decimal.Decimal) Order {
	o := newBaseOrderEvent(EventTypeFill, order)
	o.OrderStatus = EventTypeFill
	if order.LeavesQty.IsPositive() {
		o.OrderStatus = EventTypePartialFill
	}
	o.ExecQty = qty
	if tradePrice.GreaterThan(o.Price) {
		o.Price = tradePrice
	}
	o.publishEvent(o.OrderStatus)
	return o
}

func newCanceledOrderEvent(order *Order) {
	log.Printf("Creating canceled event for order: %s", order.ID)
	o := newBaseOrderEvent(EventTypeCanceled, order)
	o.LeavesQty = decimal.Zero
	o.OrderStatus = EventTypeCanceled
	o.publishEvent(EventTypeCanceled)
}

func newRejectedOrderEvent(or *OrderRequest) Order {
	o := newBaseOrder(EventTypeRejected, or.ID, or.Price, or.Side == Buy)
	o.OrderQty = or.Qty
	o.LeavesQty = decimal.Zero
	o.OrderStatus = EventTypeRejected
	o.publishEvent(EventTypeRejected)
	return o
}

// In Order struct methods where you want to publish events
func (o *Order) publishEvent(eventType EventType) {
	if o.KafkaProducer == nil {
		log.Printf("Warning: KafkaProducer is nil for order %s", o.ID)
		return
	}

	event := model.OrderEvent{
		EventType:   string(eventType),
		OrderID:     o.ID,
		Instrument:  o.Instrument,
		Price:       o.Price,
		Quantity:    o.OrderQty,
		LeavesQty:   o.LeavesQty,
		ExecQty:     o.ExecQty,
		IsBid:       o.IsBid,
		OrderStatus: string(o.OrderStatus),
		ExecType:    string(o.ExecType),
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	if err := o.KafkaProducer.NotifyEventAndTrade(o.ID, eventJSON); err != nil {
		log.Printf("Error publishing event: %v", err)
	} else {
		log.Printf("Event published for order %s: %s", o.ID, eventType)
	}
}
