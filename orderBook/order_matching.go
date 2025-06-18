package orderBook

import (
	"log"
	"sort"

	"github.com/shopspring/decimal"
)

func (book *OrderBook) processOrder(order *Order) {
	var (
		matchingBook map[decimal.Decimal]*OrderList
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
	priceKeys := make([]decimal.Decimal, 0, len(matchingBook))
	for price := range matchingBook {
		priceKeys = append(priceKeys, price)
	}

	// Prepare a sorted list of price levels from the matching side of the book
	sort.Slice(priceKeys, func(i, j int) bool {
		if isBuy {
			return priceKeys[i].LessThan(priceKeys[j]) // ascending for asks
		}
		return priceKeys[i].GreaterThan(priceKeys[j]) // descending for bids
	})

	orderMatched := false

	for _, price := range priceKeys {
		if !priceComp(price, order.Price) {
			break
		}

		orderList := matchingBook[price]
		i := 0
		for i < len(orderList.Orders) && order.LeavesQty.IsPositive() {
			match := &orderList.Orders[i]
			matchQty := decimal.Min(order.LeavesQty, match.LeavesQty)

			book.publishTrade(order, match, matchQty)

			order.LeavesQty = order.LeavesQty.Sub(matchQty)
			match.LeavesQty = match.LeavesQty.Sub(matchQty)

			NewFillOrderEvent(order, match, matchQty)

			orderMatched = true

			if match.LeavesQty.IsZero() {
				orderList.Orders = append(orderList.Orders[:i], orderList.Orders[i+1:]...)
			} else {
				i++
			}
		}

		if len(orderList.Orders) == 0 {
			delete(matchingBook, price)
		}

		if !order.LeavesQty.IsPositive() {
			break
		}
	}

	if order.LeavesQty.IsPositive() {
		book.addOrderToBook(*order)
		if !orderMatched {
			NewOrderEvent(order)
		}
	}
}

func (book *OrderBook) publishTrade(order, match *Order, qty decimal.Decimal) {
	isBuy := order.IsBid

	var buyerID, sellerID string
	if isBuy {
		buyerID = order.ID
		sellerID = match.ID
	} else {
		buyerID = match.ID
		sellerID = order.ID
	}

	price := match.Price
	if price.IsZero() {
		price = order.Price
	}

	trade := Trade{
		BuyerOrderID:  buyerID,
		SellerOrderID: sellerID,
		Quantity:      qty.BigInt().Uint64(),
		Price:         price.BigInt().Uint64(),
		Timestamp:     order.Timestamp,
		Instrument:    order.Instrument,
	}

	if book.TradeNotifier != nil {
		if err := book.TradeNotifier.NotifyEventAndTrade(trade.BuyerOrderID, trade.ToJSON()); err != nil {
			log.Printf("Error publishing event: %v", err)
		} else {
			log.Printf("Trade published BuyerId %s: SellerId %s", trade.BuyerOrderID, trade.SellerOrderID)
		}
	}
}
