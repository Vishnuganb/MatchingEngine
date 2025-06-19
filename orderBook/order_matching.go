package orderBook

import (
	"log"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/shopspring/decimal"
)

func (book *OrderBook) processOrder(order *Order) {
	var (
		matchingBook *treemap.Map
		isBuy        = order.IsBid
		priceComp    func(price decimal.Decimal) bool
	)

	if isBuy {
		matchingBook = book.Asks
		priceComp = func(price decimal.Decimal) bool {
			return price.LessThanOrEqual(order.Price)
		}
	} else {
		matchingBook = book.Bids
		priceComp = func(price decimal.Decimal) bool {
			return price.GreaterThanOrEqual(order.Price)
		}
	}

	order.LeavesQty = order.OrderQty
	it := matchingBook.Iterator()
	it.Begin()

	orderMatched := false

	for it.Next() {
		price := it.Key().(decimal.Decimal)
		if !priceComp(price) {
			break
		}

		orderList := it.Value().(*OrderList)
		i := 0
		for i < len(orderList.Orders) && order.LeavesQty.IsPositive() {
			match := &orderList.Orders[i]
			matchQty := decimal.Min(order.LeavesQty, match.LeavesQty)

			book.publishTrade(order, match, matchQty)

			order.updateOrderQuantities(matchQty)
			match.updateOrderQuantities(matchQty)

			NewFillOrderEvent(order, match, matchQty)

			orderMatched = true

			if match.LeavesQty.IsZero() {
				orderList.Orders = append(orderList.Orders[:i], orderList.Orders[i+1:]...)
			} else {
				i++
			}
		}

		if len(orderList.Orders) == 0 {
			matchingBook.Remove(price)
		}

		if !order.LeavesQty.IsPositive() {
			break
		}
	}

	if order.LeavesQty.IsPositive() {
		book.addOrderToBook(*order)
		if !orderMatched {
			order.ResetQuantities()
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
		Quantity:      qty,
		Price:         price,
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
