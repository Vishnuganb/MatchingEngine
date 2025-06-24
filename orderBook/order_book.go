package orderBook

import (
	"encoding/json"
	"log"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/shopspring/decimal"

	"MatchingEngine/internal/model"
	"MatchingEngine/internal/util"
)

type TradeNotifier interface {
	NotifyEventAndTrade(orderID string, value json.RawMessage) error
}

type OrderRef struct {
	PriceLevel decimal.Decimal
	Side       string
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

func NewOrderBook(tradeNotifier TradeNotifier) (*OrderBook, chan model.OrderRequest) {
	ob := &OrderBook{
		Bids:          treemap.NewWith(util.DecimalDescComparator),
		Asks:          treemap.NewWith(util.DecimalAscComparator),
		TradeNotifier: tradeNotifier,
		orderIndex:    make(map[string]*OrderRef),
	}
	orderChan := make(chan model.OrderRequest, 100)

	go func() {
		for req := range orderChan {
			switch req.MsgType {
			case model.MsgTypeNew:
				ob.OnNewOrder(req.NewOrderReq)
			case model.MsgTypeCancel:
				ob.CancelOrder(req.CancelOrderReq.OrigClOrdID)
			}
		}
	}()

	return ob, orderChan
}

func (book *OrderBook) OnNewOrder(or model.NewOrderRequest) {
	log.Printf("Received new order: %+v", or)
	order := convertOrderRequestToOrder(or)
	order.ExecutionNotifier = book.TradeNotifier
	order.AssignOrderID()
	err := or.ValidateNewOrder()
	if err != nil {
		log.Printf("Rejecting invalid order: %+s", err)
		order.NewRejectedOrderEvent()
		return
	}
	book.processOrder(&order)
}

func (book *OrderBook) CancelOrder(origClOrdID string) {
	ref, ok := book.orderIndex[origClOrdID]
	if !ok {
		log.Printf("Order with ID %s not found", origClOrdID)
		order := Order{ClOrdID: origClOrdID}
		order.ExecutionNotifier = book.TradeNotifier
		order.NewCanceledRejectOrderEvent()
		return
	}

	list, exists := book.getOrderListAndRemoveFromBook(ref.PriceLevel, ref.Side == string(model.Buy))

	if !exists || list == nil || ref.Index >= len(list.Orders) || list.Orders[ref.Index].ClOrdID != origClOrdID {
		log.Printf("Order with ID %s inconsistent in index", origClOrdID)
		delete(book.orderIndex, origClOrdID)
		order := Order{ClOrdID: origClOrdID}
		order.ExecutionNotifier = book.TradeNotifier
		order.NewCanceledRejectOrderEvent()
		return
	}

	order := list.Orders[ref.Index]
	order.ExecutionNotifier = book.TradeNotifier

	// Remove order from list (swap with last, then pop)
	last := len(list.Orders) - 1
	if ref.Index != last {
		list.Orders[ref.Index] = list.Orders[last]
		// Update moved order's index
		book.orderIndex[list.Orders[ref.Index].ClOrdID].Index = ref.Index
	}
	list.Orders = list.Orders[:last]

	if len(list.Orders) == 0 {
		if ref.Side == string(model.Buy) {
			book.Bids.Remove(ref.PriceLevel)
		} else {
			book.Asks.Remove(ref.PriceLevel)
		}
	}

	delete(book.orderIndex, origClOrdID)

	log.Printf("Canceled order %s from %s at price %s", origClOrdID, ref.Side, ref.PriceLevel)
	order.NewCanceledOrderEvent()
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

	if order.Side == model.Buy {
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
	book.orderIndex[order.ClOrdID] = &OrderRef{
		PriceLevel: priceKey,
		Side:       string(order.Side),
		Index:      len(list.Orders) - 1,
	}
}

func convertOrderRequestToOrder(or model.NewOrderRequest) Order {
	return Order{
		ClOrdID:     or.ClOrdID,
		Symbol:      or.Symbol,
		Side:        or.Side,
		Price:       or.Price,
		OrderQty:    or.OrderQty,
		LeavesQty:   or.OrderQty,
		Timestamp:   or.TransactTime,
		Text:        or.Text,
		OrderStatus: model.OrderStatusPendingNew,
		CumQty:      decimal.Zero,
	}
}
