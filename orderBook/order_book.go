package orderBook

import (
	"encoding/json"
	"log"

	"github.com/shopspring/decimal"

	"MatchingEngine/internal/model"
)

type EventNotifier interface {
	NotifyEventAndTrade(orderID string, value json.RawMessage) error
}

type OrderBook struct {
	Bids          []Order `json:"bids"`
	Asks          []Order `json:"asks"`
	Orders        []Order `json:"orders"`
	KafkaProducer EventNotifier
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Bids:   []Order{},
		Asks:   []Order{},
		Orders: []Order{},
	}
}

func (book *OrderBook) OnNewOrder(modelOrder model.Order) model.Orders {
	var trades []Trade
	var orders Orders
	order := mapModelOrderToOrderBookOrder(modelOrder)
	if order.IsBid {
		trades = book.processBuyOrder(&order)
		log.Println("Trades", trades)
	} else {
		trades = book.processSellOrder(&order)
	}
	// Emit fill events for each trade
	if len(trades) > 0 {
		for _, trade := range trades {
			fillQty := decimal.NewFromInt(int64(trade.Quantity))
			price := decimal.NewFromInt(int64(trade.Price))
			orders = append(orders, newFillOrderEvent(&order,
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
		newOrder := newOrderEvent(&order)
		orders = append(orders, newOrder)
	}

	book.Orders = append(book.Orders, orders...)
	return mapOrderBookOrdersToModelOrders(orders)
}

func (book *OrderBook) CancelOrder(orderID string) model.Order {
	// Search for the order in bids
	for i, order := range book.Bids {
		if order.ID == orderID {
			book.RemoveBuyOrder(i)
			orderEvent := newCanceledOrderEvent(&order)
			return mapOrderBookOrderToModelOrder(orderEvent)
		}
	}

	// Search for the order in asks
	for i, order := range book.Asks {
		if order.ID == orderID {
			book.RemoveSellOrder(i)
			orderEvent := newCanceledOrderEvent(&order)
			return mapOrderBookOrderToModelOrder(orderEvent)
		}
	}

	// If the order is not found, return an empty event
	return model.Order{}
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

func mapOrderBookOrderToModelOrder(order Order) model.Order {
	return model.Order{
		ID:          order.ID,
		Instrument:  order.Instrument,
		Timestamp:   order.Timestamp,
		ExecType:    string(order.ExecType),
		IsBid:       order.IsBid,
		Price:       order.Price,
		OrderQty:    order.OrderQty,
		LeavesQty:   order.LeavesQty,
		ExecQty:     order.ExecQty,
		OrderStatus: string(order.OrderStatus),
	}
}

func mapOrderBookOrdersToModelOrders(orders Orders) model.Orders {
	mappedEvents := make([]model.Order, len(orders))
	for i, order := range orders {
		mappedEvents[i] = model.Order{
			ID:          order.ID,
			Instrument:  order.Instrument,
			Timestamp:   order.Timestamp,
			ExecType:    string(order.ExecType),
			IsBid:       order.IsBid,
			Price:       order.Price,
			OrderQty:    order.OrderQty,
			LeavesQty:   order.LeavesQty,
			ExecQty:     order.ExecQty,
			OrderStatus: string(order.OrderStatus),
		}
	}
	return mappedEvents
}

func mapModelOrderToOrderBookOrder(order model.Order) Order {
	return Order{
		ID:          order.ID,
		Instrument:  order.Instrument,
		Timestamp:   order.Timestamp,
		ExecType:    EventType(order.ExecType),
		IsBid:       order.IsBid,
		Price:       order.Price,
		OrderQty:    order.OrderQty,
		LeavesQty:   order.LeavesQty,
		ExecQty:     order.ExecQty,
		OrderStatus: EventType(order.OrderStatus),
	}
}
