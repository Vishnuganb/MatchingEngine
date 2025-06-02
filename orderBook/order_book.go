package orderBook

import "github.com/shopspring/decimal"

type OrderBook struct {
	Bids   []Order `json:"bids"`
	Asks   []Order `json:"asks"`
	Events []Event `json:"events"`
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Bids:   []Order{},
		Asks:   []Order{},
		Events: []Event{},
	}
}

func (book *OrderBook) NewOrder(order Order) Event {
	var trade Trade
	if order.IsBid {
		trade = book.processBuyOrder(order)
	} else {
		trade = book.processSellOrder(order)
	}

	// Generate an event for the trade
	if trade.Quantity > 0 {
		return newFillEvent(&order, decimal.NewFromInt(int64(trade.Quantity)))
	}

	// If the order is not fully filled, add it to the order book
	if order.LeavesQty.IsPositive() {
		if order.IsBid {
			book.AddBuyOrder(order)
		} else {
			book.AddSellOrder(order)
		}
	}

	return newEvent(&order)
}

func (book *OrderBook) CancelOrder(orderID string) Event {
	// Search for the order in bids
	for i, order := range book.Bids {
		if order.ID == orderID {
			book.RemoveBuyOrder(i)
			return newCanceledEvent(&order)
		}
	}

	// Search for the order in asks
	for i, order := range book.Asks {
		if order.ID == orderID {
			book.RemoveSellOrder(i)
			return newCanceledEvent(&order)
		}
	}

	// If the order is not found, return an empty event
	return Event{}
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
