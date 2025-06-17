package orderBook

import (
	"encoding/json"
	"log"

	"github.com/shopspring/decimal"
)

type TradeNotifier interface {
	NotifyEventAndTrade(orderID string, value json.RawMessage) error
}

type OrderRef struct {
	PriceLevel decimal.Decimal
	IsBid      bool
	Index      int
}

type OrderBook struct {
	Bids          map[decimal.Decimal]*OrderList // descending prices
	Asks          map[decimal.Decimal]*OrderList // ascending prices
	TradeNotifier TradeNotifier
	orderIndex    map[string]*OrderRef
}

type OrderList struct {
	Orders []Order
}

func NewOrderBook(tradeNotifier TradeNotifier) *OrderBook {
	return &OrderBook{
		Bids:          make(map[decimal.Decimal]*OrderList),
		Asks:          make(map[decimal.Decimal]*OrderList),
		TradeNotifier: tradeNotifier,
		orderIndex:    make(map[string]*OrderRef),
	}
}

func (book *OrderBook) OnNewOrder(order Order) {
	if !order.OrderQty.IsPositive() || order.Price.IsNegative() {
		log.Printf("Rejecting invalid order: %+v", order)
		NewRejectedOrderEvent(&OrderRequest{
			ID:        order.ID,
			Price:     order.Price,
			Qty:       order.OrderQty,
			Side:      ternary(order.IsBid, Buy, Sell),
			Timestamp: order.Timestamp,
		})
		return
	}

	order.ExecutionNotifier = book.TradeNotifier
	book.processOrder(&order)
}

func (book *OrderBook) CancelOrder(orderID string) {
	ref, ok := book.orderIndex[orderID]
	if !ok {
		log.Printf("Order with ID %s not found", orderID)
		return
	}

	var list *OrderList
	if ref.IsBid {
		list = book.Bids[ref.PriceLevel]
	} else {
		list = book.Asks[ref.PriceLevel]
	}

	if list == nil || ref.Index >= len(list.Orders) || list.Orders[ref.Index].ID != orderID {
		log.Printf("Order with ID %s inconsistent in index", orderID)
		delete(book.orderIndex, orderID)
		return
	}

	order := list.Orders[ref.Index]
	order.ExecutionNotifier = book.TradeNotifier

	// Remove order from list (swap with last, then pop)
	last := len(list.Orders) - 1
	if ref.Index != last {
		list.Orders[ref.Index] = list.Orders[last]
		// Update moved order's index
		book.orderIndex[list.Orders[ref.Index].ID].Index = ref.Index
	}
	list.Orders = list.Orders[:last]

	if len(list.Orders) == 0 {
		if ref.IsBid {
			delete(book.Bids, ref.PriceLevel)
		} else {
			delete(book.Asks, ref.PriceLevel)
		}
	}

	delete(book.orderIndex, orderID)

	log.Printf("Canceled order %s from %s at price %s", orderID, ternary(ref.IsBid, "Bids", "Asks"), ref.PriceLevel)
	NewCanceledOrderEvent(&order)
}

func (book *OrderBook) addOrderToBook(order Order) {
	priceKey := order.Price
	var list *OrderList

	if order.IsBid {
		if _, ok := book.Bids[priceKey]; !ok {
			book.Bids[priceKey] = &OrderList{}
		}
		list = book.Bids[priceKey]
	} else {
		if _, ok := book.Asks[priceKey]; !ok {
			book.Asks[priceKey] = &OrderList{}
		}
		list = book.Asks[priceKey]
	}

	list.Orders = append(list.Orders, order)
	book.orderIndex[order.ID] = &OrderRef{
		PriceLevel: priceKey,
		IsBid:      order.IsBid,
		Index:      len(list.Orders) - 1,
	}
}

func ternary(condition bool, a, b Side) Side {
	if condition {
		return a
	}
	return b
}
