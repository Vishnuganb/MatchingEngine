package orderBook

import (
	"encoding/json"
	"log"

	"github.com/shopspring/decimal"
)

func (book *OrderBook) processOrder(order *Order, isBuy bool) []Trade {
	log.Println("processOrder")

	var matchingOrders *[]Order
	var removeOrderFunc func(int)
	var priceComparison func(decimal.Decimal, decimal.Decimal) bool

	if isBuy {
		matchingOrders = &book.Asks
		removeOrderFunc = book.RemoveSellOrder
		priceComparison = func(a, b decimal.Decimal) bool { return a.LessThanOrEqual(b) }
	} else {
		matchingOrders = &book.Bids
		removeOrderFunc = book.RemoveBuyOrder
		priceComparison = func(a, b decimal.Decimal) bool { return a.GreaterThanOrEqual(b) }
	}

	// Validate order
	if !order.OrderQty.IsPositive() || order.Price.IsNegative() {
		book.Orders = append(book.Orders, newRejectedOrderEvent(&OrderRequest{
			ID:   order.ID,
			Side: order.Side(),
			Qty:  order.OrderQty,
		}))
		return nil
	}

	var trades []Trade
	order.LeavesQty = order.OrderQty

	// Try to match against the best available orders
	for len(*matchingOrders) > 0 && order.LeavesQty.IsPositive() {
		matchingOrder := (*matchingOrders)[0]

		if !priceComparison(matchingOrder.Price, order.Price) {
			break // No more matching prices
		}

		// Calculate trade quantity
		matchQty := decimal.Min(order.LeavesQty, matchingOrder.LeavesQty)

		trade := Trade{
			BuyerOrderID:  getOrderID(order, &matchingOrder, isBuy),
			SellerOrderID: getOrderID(order, &matchingOrder, !isBuy),
			Quantity:      matchQty.BigInt().Uint64(),
			Price:         matchingOrder.Price.BigInt().Uint64(),
			Timestamp:     order.Timestamp,
		}

		book.publishTrade(trade)
		trades = append(trades, trade)

		// Update LeavesQty
		order.LeavesQty = order.LeavesQty.Sub(matchQty)
		matchingOrder.LeavesQty = matchingOrder.LeavesQty.Sub(matchQty)

		if matchingOrder.LeavesQty.IsZero() {
			removeOrderFunc(0)
		} else {
			(*matchingOrders)[0] = matchingOrder
			break // Cannot consume further from this price level
		}
	}

	// No match: add to order book
	return trades
}

func (book *OrderBook) processBuyOrder(order *Order) []Trade {
	return book.processOrder(order, true)
}

func (book *OrderBook) processSellOrder(order *Order) []Trade {
	return book.processOrder(order, false)
}

func getOrderID(order, matchingOrder *Order, isBuy bool) string {
	if isBuy {
		return order.ID
	}
	return matchingOrder.ID
}

func (book *OrderBook) publishTrade(trade Trade) {
	if book.KafkaProducer != nil {
		err := book.KafkaProducer.NotifyEventAndTrade(trade.BuyerOrderID, json.RawMessage(trade.ToJSON()))
		if err != nil {
			return
		}
	}
}
