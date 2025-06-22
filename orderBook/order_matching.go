package orderBook

import (
	"MatchingEngine/internal/model"
	"MatchingEngine/internal/util"
	"log"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/shopspring/decimal"
)

func (book *OrderBook) processOrder(order *Order) {
	var (
		matchingBook *treemap.Map
		isBuy        = order.Side == model.Buy
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
			order.NewOrderEvent()
		}
	}
}

func (book *OrderBook) publishTrade(order, match *Order, qty decimal.Decimal) {
	isBuy := order.Side == model.Buy

	var (
		side model.Side
	)

	if isBuy {
		side = model.Buy
	} else {
		side = model.Sell
	}

	price := match.Price
	if price.IsZero() {
		price = order.Price
	}

	tradeReport := model.TradeCaptureReport{
		MsgType:            "AE",                                   // FIX MsgType = AE (Trade Capture Report)
		TradeReportID:      util.GeneratePrefixedID("tradeReport"), // Unique trade report ID
		OrderID:            order.OrderID,                          // Exchange-assigned order ID
		ClOrdID:            order.ClOrdID,                          // Client Order ID
		Symbol:             order.Symbol,
		Side:               side,
		LastQty:            qty,
		LastPx:             price,
		TradeDate:          util.FormatDate(order.Timestamp), // Format: YYYYMMDD
		TransactTime:       order.Timestamp,
		PreviouslyReported: false,
	}

	if book.TradeNotifier != nil {
		if err := book.TradeNotifier.NotifyEventAndTrade(tradeReport.TradeReportID, tradeReport.ToJSON()); err != nil {
			log.Printf("Error publishing trade report: %v", err)
		} else {
			//log.Printf("Trade report sent: ExecID %s, Buy=%s Sell=%s Qty=%s",
			//	trade.ExecID, buyer.ClOrdID, seller.ClOrdID, qty.String())
		}
	}
}
