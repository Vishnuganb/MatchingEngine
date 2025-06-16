package orderBook

import (
	"encoding/json"
	"log"

	"MatchingEngine/internal/model"
)

type TradeNotifier interface {
	NotifyEventAndTrade(orderID string, value json.RawMessage) error
}

type OrderRef struct {
	PriceLevel string
	IsBid      bool
	Index      int
}

type OrderBook struct {
	Bids          map[string]*OrderList // descending prices
	Asks          map[string]*OrderList // ascending prices
	TradeNotifier TradeNotifier
	orderIndex    map[string]*OrderRef
}

type OrderList struct {
	Orders []Order
}

func NewOrderBook(tradeNotifier TradeNotifier) *OrderBook {
	return &OrderBook{
		Bids:          make(map[string]*OrderList),
		Asks:          make(map[string]*OrderList),
		TradeNotifier: tradeNotifier,
		orderIndex:    make(map[string]*OrderRef),
	}
}

func (book *OrderBook) OnNewOrder(modelOrder model.Order) {
	// Basic validation
	if !modelOrder.OrderQty.IsPositive() || modelOrder.Price.IsNegative() {
		log.Printf("Rejecting invalid order: %+v", modelOrder)
		NewRejectedOrderEvent(&OrderRequest{
			ID:        modelOrder.ID,
			Price:     modelOrder.Price,
			Qty:       modelOrder.OrderQty,
			Side:      ternary(modelOrder.IsBid, Buy, Sell),
			Timestamp: modelOrder.Timestamp,
		})
		return
	}

	order := mapModelOrderToOrderBookOrder(modelOrder)
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

	// Remove order from list efficiently (swap with last, then pop)
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
	priceKey := order.Price.String()
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

func mapModelOrderToOrderBookOrder(order model.Order) Order {
	return Order{
		ID:          order.ID,
		Instrument:  order.Instrument,
		Timestamp:   order.Timestamp,
		IsBid:       order.IsBid,
		Price:       order.Price,
		OrderQty:    order.OrderQty,
		LeavesQty:   order.OrderQty, // Assume full qty on entry
		CumQty:      order.CumQty,
		OrderStatus: OrderStatus(order.OrderStatus),
	}
}

func ternary(condition bool, a, b Side) Side {
	if condition {
		return a
	}
	return b
}
