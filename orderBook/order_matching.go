package orderBook

import (
	"log"

	"github.com/shopspring/decimal"
)

func (book *OrderBook) processOrder(order Order, isBuy bool) Trade {
	log.Println("processOrder")

	var matchingOrders *[]Order
	var addOrderFunc func(Order)
	var removeOrderFunc func(int)
	var priceComparison func(decimal.Decimal, decimal.Decimal) bool

	if isBuy {
		matchingOrders = &book.Asks
		addOrderFunc = book.AddBuyOrder
		removeOrderFunc = book.RemoveSellOrder
		priceComparison = func(a, b decimal.Decimal) bool { return a.LessThanOrEqual(b) }
	} else {
		matchingOrders = &book.Bids
		addOrderFunc = book.AddSellOrder
		removeOrderFunc = book.RemoveBuyOrder
		priceComparison = func(a, b decimal.Decimal) bool { return a.GreaterThanOrEqual(b) }
	}

	// Validate order
	if !order.Qty.IsPositive() || order.Price.IsNegative() {
		book.Events = append(book.Events, newRejectedEvent(&OrderRequest{
			ID:   order.ID,
			Side: order.Side(),
			Qty:  order.Qty,
		}))
		return Trade{}
	}

	// Try to match against best available orders
	for len(*matchingOrders) > 0 {
		matchingOrder := (*matchingOrders)[0]

		if !priceComparison(matchingOrder.Price, order.Price) {
			break // No more matching prices
		}

		if matchingOrder.Qty.GreaterThanOrEqual(order.Qty) {
			// Full match
			trade := Trade{
				BuyerOrderID:  getOrderID(order, isBuy),
				SellerOrderID: getOrderID(matchingOrder, !isBuy),
				Quantity:      order.Qty.BigInt().Uint64(),
				Price:         matchingOrder.Price.BigInt().Uint64(),
				Timestamp:     order.Timestamp,
			}

			// Update remaining qty of matching order
			matchingOrder.LeavesQty = matchingOrder.LeavesQty.Sub(order.Qty)
			if matchingOrder.LeavesQty.IsZero() {
				removeOrderFunc(0)
			} else {
				(*matchingOrders)[0] = matchingOrder
			}
			return trade
		} else {
			// Partial match
			trade := Trade{
				BuyerOrderID:  getOrderID(order, isBuy),
				SellerOrderID: getOrderID(matchingOrder, !isBuy),
				Quantity:      matchingOrder.Qty.BigInt().Uint64(),
				Price:         matchingOrder.Price.BigInt().Uint64(),
				Timestamp:     order.Timestamp,
			}

			order.Qty = order.Qty.Sub(matchingOrder.Qty)
			removeOrderFunc(0) // Remove the fully matched order
			return trade
		}
	}

	// No match: add to order book
	addOrderFunc(order)
	book.Events = append(book.Events, newEvent(&order))
	return Trade{}
}

func (book *OrderBook) processBuyOrder(order Order) Trade {
	return book.processOrder(order, true)
}

func (book *OrderBook) processSellOrder(order Order) Trade {
	return book.processOrder(order, false)
}

func getOrderID(order Order, isBuy bool) string {
	if isBuy {
		return order.ID
	}
	return order.ID
}
