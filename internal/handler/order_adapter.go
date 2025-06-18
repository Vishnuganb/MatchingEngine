package handler

import (
	"time"

	"github.com/shopspring/decimal"

	"MatchingEngine/internal/rmq"
	"MatchingEngine/orderBook"
)

func toInternalOrder(req rmq.OrderRequest) orderBook.OrderRequest {
	qty := decimal.RequireFromString(req.Order.Qty)
	return orderBook.OrderRequest{
		ID:          req.Order.ID,
		Instrument:  req.Order.Instrument,
		Timestamp:   time.Now().UnixNano(),
		Side:        req.Order.Side,
		Price:       decimal.RequireFromString(req.Order.Price),
		Qty:         qty,
	}
}
