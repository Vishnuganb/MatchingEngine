package orderBook

import (
	"log"
	"sort"

	"github.com/shopspring/decimal"
)

func (book *OrderBook) processOrder(order *Order) {
	var (
		matchingBook map[string]*OrderList
		isBuy        = order.IsBid
		priceComp    func(decimal.Decimal, decimal.Decimal) bool
	)

	if isBuy {
		matchingBook = book.Asks
		priceComp = func(bookPrice, orderPrice decimal.Decimal) bool {
			return bookPrice.LessThanOrEqual(orderPrice)
		}
	} else {
		matchingBook = book.Bids
		priceComp = func(bookPrice, orderPrice decimal.Decimal) bool {
			return bookPrice.GreaterThanOrEqual(orderPrice)
		}
	}

	order.LeavesQty = order.OrderQty
	// collecting all the price levels from the matchingBook
	priceKeys := make([]string, 0, len(matchingBook))
	for price := range matchingBook {
		priceKeys = append(priceKeys, price)
	}

	// Prepare a sorted list of price levels from the matching side of the book
	sort.Slice(priceKeys, func(i, j int) bool {
		pi := decimal.RequireFromString(priceKeys[i])
		pj := decimal.RequireFromString(priceKeys[j])
		if isBuy {
			return pi.LessThan(pj) // ascending for asks
		}
		return pi.GreaterThan(pj) // descending for bids
	})

	for _, priceStr := range priceKeys {
		price := decimal.RequireFromString(priceStr)
		if !priceComp(price, order.Price) {
			break
		}

		orderList := matchingBook[priceStr]
		i := 0
		for i < len(orderList.Orders) && order.LeavesQty.IsPositive() {
			match := &orderList.Orders[i]
			matchQty := decimal.Min(order.LeavesQty, match.LeavesQty)

			trade := Trade{
				BuyerOrderID:  getOrderID(order, match, isBuy),
				SellerOrderID: getOrderID(order, match, !isBuy),
				Quantity:      matchQty.BigInt().Uint64(),
				Price:         price.BigInt().Uint64(),
				Timestamp:     order.Timestamp,
			}

			book.publishTrade(trade)

			order.LeavesQty = order.LeavesQty.Sub(matchQty)
			match.LeavesQty = match.LeavesQty.Sub(matchQty)

			newFillOrderEvent(order, matchQty, price)
			newFillOrderEvent(match, matchQty, price)

			if match.LeavesQty.IsZero() {
				orderList.Orders = append(orderList.Orders[:i], orderList.Orders[i+1:]...)
			} else {
				i++
			}
		}

		if len(orderList.Orders) == 0 {
			delete(matchingBook, priceStr)
		}

		if !order.LeavesQty.IsPositive() {
			break
		}
	}

	if order.LeavesQty.IsPositive() {
		book.addOrderToBook(*order)
	}
}

func getOrderID(order, matchingOrder *Order, isBuy bool) string {
	if isBuy {
		return order.ID
	}
	return matchingOrder.ID
}

func (book *OrderBook) publishTrade(trade Trade) {
	if book.TradeNotifier != nil {
		if err := book.TradeNotifier.NotifyEventAndTrade(trade.BuyerOrderID, trade.ToJSON()); err != nil {
			log.Printf("Error publishing event: %v", err)
		} else {
			log.Printf("Trade published BuyerId %s: SellerId %s", trade.BuyerOrderID, trade.SellerOrderID)
		}
	}
}
