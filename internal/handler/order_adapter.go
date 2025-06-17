package handler

import (
	"time"

	"github.com/shopspring/decimal"

	"MatchingEngine/internal/rmq"
	"MatchingEngine/orderBook"
)

func toInternalOrder(req rmq.OrderRequest) orderBook.Order {
	qty := decimal.RequireFromString(req.Order.Qty)
	return orderBook.Order{
		ID:          req.Order.ID,
		Instrument:  req.Order.Instrument,
		Timestamp:   time.Now().UnixNano(),
		IsBid:       req.Order.Side == orderBook.Buy,
		Price:       decimal.RequireFromString(req.Order.Price),
		OrderQty:    qty,
		LeavesQty:   qty,
		OrderStatus: orderBook.OrderStatusPendingNew,
	}
}
