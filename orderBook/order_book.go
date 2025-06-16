package orderBook

import (
	"encoding/json"
	"log"

	"MatchingEngine/internal/model"
)

type TradeNotifier interface {
	NotifyEventAndTrade(orderID string, value json.RawMessage) error
}

type OrderBook struct {
	Bids          map[string]*OrderList // descending prices
	Asks          map[string]*OrderList // ascending prices
	TradeNotifier TradeNotifier
}

type OrderList struct {
	Orders []Order
}

func NewOrderBook(tradeNotifier TradeNotifier) *OrderBook {
	return &OrderBook{
		Bids:          make(map[string]*OrderList),
		Asks:          make(map[string]*OrderList),
		TradeNotifier: tradeNotifier,
	}
}

func (book *OrderBook) OnNewOrder(modelOrder model.Order) {
	order := mapModelOrderToOrderBookOrder(modelOrder)
	order.ExecutionNotifier = book.TradeNotifier
	book.processOrder(&order)
	NewOrderEvent(&order)
}

func (book *OrderBook) CancelOrder(orderID string) {
	for price, list := range book.Bids {
		for i, order := range list.Orders {
			if order.ID == orderID {
				order.ExecutionNotifier = book.TradeNotifier
				list.Orders = append(list.Orders[:i], list.Orders[i+1:]...)
				log.Printf("Canceled order %s from Bids at price %s", orderID, price)
				NewCanceledOrderEvent(&order)
				return
			}
		}
	}

	for price, list := range book.Asks {
		for i, order := range list.Orders {
			if order.ID == orderID {
				order.ExecutionNotifier = book.TradeNotifier
				list.Orders = append(list.Orders[:i], list.Orders[i+1:]...)
				log.Printf("Canceled order %s from Asks at price %s", orderID, price)
				NewCanceledOrderEvent(&order)
				return
			}
		}
	}

	log.Printf("Order with ID %s not found", orderID)
}

func (book *OrderBook) addOrderToBook(order Order) {
	priceKey := order.Price.String()
	if order.IsBid {
		if _, ok := book.Bids[priceKey]; !ok {
			book.Bids[priceKey] = &OrderList{}
		}
		book.Bids[priceKey].Orders = append(book.Bids[priceKey].Orders, order)
	} else {
		if _, ok := book.Asks[priceKey]; !ok {
			book.Asks[priceKey] = &OrderList{}
		}
		book.Asks[priceKey].Orders = append(book.Asks[priceKey].Orders, order)
	}
}

func mapModelOrderToOrderBookOrder(order model.Order) Order {
	return Order{
		ID:          order.ID,
		Instrument:  order.Instrument,
		Timestamp:   order.Timestamp,
		ExecType:    ExecType(order.ExecType),
		IsBid:       order.IsBid,
		Price:       order.Price,
		OrderQty:    order.OrderQty,
		LeavesQty:   order.OrderQty, // Assume full qty on entry
		CumQty:      order.ExecQty,
		OrderStatus: OrderStatus(order.OrderStatus),
	}
}
