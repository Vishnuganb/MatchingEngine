package orderBook

import (
	"MatchingEngine/internal/kafka"
	"MatchingEngine/internal/model"
	"github.com/shopspring/decimal"
	"log"
)

type OrderBook struct {
	Bids   []Order `json:"bids"`
	Asks   []Order `json:"asks"`
	Events []Event `json:"events"`
	KafkaProducer kafka.EventNotifier
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Bids:   []Order{},
		Asks:   []Order{},
		Events: []Event{},
	}
}

func (book *OrderBook) OnNewOrder(modelOrder model.Order) model.Events {
	var trades []Trade
	var events Events
	order := mapModelOrderToOrderBookOrder(modelOrder)
	if order.IsBid {
		trades = book.processBuyOrder(&order)
		log.Println("orderbook", trades)
	} else {
		trades = book.processSellOrder(&order)
	}
	// Emit fill events for each trade
	if len(trades) > 0 {
		for _, trade := range trades {
			fillQty := decimal.NewFromInt(int64(trade.Quantity))
			price := decimal.NewFromInt(int64(trade.Price))
			events = append(events, newFillEvent(&order,
				fillQty, price),
			)
		}
	}

	// If the order is not fully filled, add it to the order book
	if order.LeavesQty.IsPositive() {
		if order.IsBid {
			book.AddBuyOrder(order)
		} else {
			book.AddSellOrder(order)
		}
		events = append(events, newEvent(&order))
	}

	book.Events = append(book.Events, events...)
	return mapOrderBookEventsToModelEvents(events)
}

func (book *OrderBook) CancelOrder(orderID string) model.Event {
	// Search for the order in bids
	for i, order := range book.Bids {
		if order.ID == orderID {
			book.RemoveBuyOrder(i)
			return mapOrderBookEventToModelEvent(newCanceledEvent(&order))
		}
	}

	// Search for the order in asks
	for i, order := range book.Asks {
		if order.ID == orderID {
			book.RemoveSellOrder(i)
			return mapOrderBookEventToModelEvent(newCanceledEvent(&order))
		}
	}

	// If the order is not found, return an empty event
	return model.Event{}
}

// Add the new Order to end of orderbook in bids
func (book *OrderBook) AddBuyOrder(order Order) {
	n := len(book.Bids)

	if n == 0 {
		book.Bids = append(book.Bids, order)
	} else {
		var i int

		for i := n - 1; i >= 0; i-- {
			buyOrder := book.Bids[i]
			if buyOrder.Price.LessThan(order.Price) {
				break
			}
		}

		if i == n-1 {
			book.Bids = append(book.Bids, order)
		} else {
			book.Bids = append(book.Bids, Order{})
			copy(book.Bids[i+1:], book.Bids[i:])
			book.Bids[i] = order
		}
	}
}

// Add the new Order to end of orderbook in asks
func (book *OrderBook) AddSellOrder(order Order) {
	n := len(book.Asks)

	if n == 0 {
		book.Asks = append(book.Asks, order)
	} else {
		var i int

		for i := n - 1; i >= 0; i-- {
			sellOrder := book.Asks[i]
			if sellOrder.Price.GreaterThan(order.Price) {
				break
			}
		}
		if i == n-1 {
			book.Asks = append(book.Asks, order)
		} else {
			book.Asks = append(book.Asks, Order{})
			copy(book.Asks[i+1:], book.Asks[i:])
			book.Asks[i] = order
		}
	}
}

func (book *OrderBook) RemoveBuyOrder(index int) {
	book.Bids = append(book.Bids[:index], book.Bids[index+1:]...)
}

func (book *OrderBook) RemoveSellOrder(index int) {
	book.Asks = append(book.Asks[:index], book.Asks[index+1:]...)
}

func mapOrderBookEventToModelEvent(event Event) model.Event {
	return model.Event{
		ID:         event.ID,
		OrderID:    event.OrderID,
		Instrument: event.Instrument,
		Timestamp:  event.Timestamp,
		Type:       string(event.Type),
		Side:       string(event.Side),
		Price:      event.Price,
		OrderQty:   event.OrderQty,
		LeavesQty:  event.LeavesQty,
		ExecQty:    event.ExecQty,
	}
}

func mapOrderBookEventsToModelEvents(events Events) model.Events {
	mappedEvents := make([]model.Event, len(events))
	for i, event := range events {
		mappedEvents[i] = model.Event{
			ID:         event.ID,
			OrderID:    event.OrderID,
			Instrument: event.Instrument,
			Timestamp:  event.Timestamp,
			Type:       string(event.Type),
			Side:       string(event.Side),
			Price:      event.Price,
			OrderQty:   event.OrderQty,
			LeavesQty:  event.LeavesQty,
			ExecQty:    event.ExecQty,
		}
	}
	return mappedEvents
}

func mapModelOrderToOrderBookOrder(order model.Order) Order {
	return Order{
		ID:         order.ID,
		Instrument: order.Instrument,
		Price:      order.Price,
		Qty:        order.Qty,
		LeavesQty:  order.LeavesQty,
		Timestamp:  order.Timestamp,
		IsBid:      order.IsBid,
	}
}
