package orderBook

import (
	"encoding/json"
	"log"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/shopspring/decimal"

	"MatchingEngine/internal/util"
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
	Bids          *treemap.Map // sorted descending // descending prices
	Asks          *treemap.Map // ascending prices
	TradeNotifier TradeNotifier
	orderIndex    map[string]*OrderRef
}

type OrderList struct {
	Orders []Order
}

func NewOrderBook(tradeNotifier TradeNotifier) *OrderBook {
	return &OrderBook{
		Bids:          treemap.NewWith(util.DecimalDescComparator),
		Asks:          treemap.NewWith(util.DecimalAscComparator),
		TradeNotifier: tradeNotifier,
		orderIndex:    make(map[string]*OrderRef),
	}
}

func (book *OrderBook) OnNewOrder(or OrderRequest) {
	order := convertOrderRequestToOrder(or)
	order.ExecutionNotifier = book.TradeNotifier
	err := or.Validate()
	if err != nil {
		log.Printf("Rejecting invalid order: %+s", err)
		NewRejectedOrderEvent(&order)
		return
	}
	book.processOrder(&order)
}

func (book *OrderBook) CancelOrder(orderID string) {
	ref, ok := book.orderIndex[orderID]
	if !ok {
		log.Printf("Order with ID %s not found", orderID)
		NewCanceledRejectOrderEvent(&Order{ID: orderID})
		return
	}

	list, exists := book.getOrderListAndRemoveFromBook(ref.PriceLevel, ref.IsBid)

	if !exists || list == nil || ref.Index >= len(list.Orders) || list.Orders[ref.Index].ID != orderID {
		log.Printf("Order with ID %s inconsistent in index", orderID)
		delete(book.orderIndex, orderID)
		NewCanceledRejectOrderEvent(&Order{ID: orderID})
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
			book.Bids.Remove(ref.PriceLevel)
		} else {
			book.Asks.Remove(ref.PriceLevel)
		}
	}

	delete(book.orderIndex, orderID)

	log.Printf("Canceled order %s from %s at price %s", orderID, ternary(ref.IsBid, "Bids", "Asks"), ref.PriceLevel)
	NewCanceledOrderEvent(&order)
}

func (book *OrderBook) getOrderListAndRemoveFromBook(priceLevel decimal.Decimal, isBid bool) (*OrderList, bool) {
	var list *OrderList
	var ok bool

	if isBid {
		val, exists := book.Bids.Get(priceLevel)
		if exists {
			list = val.(*OrderList)
			ok = true
		} else {
			log.Printf("Order with ID not found in Bids at price %s", priceLevel)
			book.Bids.Remove(priceLevel)
		}
	} else {
		val, exists := book.Asks.Get(priceLevel)
		if exists {
			list = val.(*OrderList)
			ok = true
		} else {
			log.Printf("Order with ID not found in Asks at price %s", priceLevel)
			book.Asks.Remove(priceLevel)
		}
	}

	return list, ok
}

func (book *OrderBook) addOrderToBook(order Order) {
	priceKey := order.Price
	var list *OrderList

	if order.IsBid {
		if val, ok := book.Bids.Get(priceKey); !ok {
			list = &OrderList{}
			book.Bids.Put(priceKey, list)
		} else {
			list = val.(*OrderList)
		}
	} else {
		if val, ok := book.Asks.Get(priceKey); !ok {
			list = &OrderList{}
			book.Asks.Put(priceKey, list)
		} else {
			list = val.(*OrderList)
		}
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

func convertOrderRequestToOrder(or OrderRequest) Order {
	return Order{
		ID:          or.ID,
		Instrument:  or.Instrument,
		Price:       or.Price,
		OrderQty:    or.Qty,
		LeavesQty:   or.Qty,
		Timestamp:   or.Timestamp,
		IsBid:       or.Side == "buy",
		OrderStatus: OrderStatusPendingNew,
		CumQty:      decimal.Zero,
	}
}
