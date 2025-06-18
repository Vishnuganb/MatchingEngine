package handler

import (
	"time"

	"github.com/shopspring/decimal"

	"MatchingEngine/internal/model"
	"MatchingEngine/internal/rmq"
	"MatchingEngine/orderBook"
)

func toInternalOrderRequest(req rmq.OrderRequest) orderBook.OrderRequest {
	qty := decimal.RequireFromString(req.Order.Qty)
	return orderBook.OrderRequest{
		ID:         req.Order.ID,
		Instrument: req.Order.Instrument,
		Timestamp:  time.Now().UnixNano(),
		Side:       req.Order.Side,
		Price:      decimal.RequireFromString(req.Order.Price),
		Qty:        qty,
	}
}

func convertEventToOrder(execution model.ExecutionReport) model.Order {
	return model.Order{
		ID:          execution.OrderID,
		Instrument:  execution.Instrument,
		Price:       execution.Price,
		OrderQty:    execution.OrderQty,
		LeavesQty:   execution.LeavesQty,
		CumQty:      execution.CumQty,
		IsBid:       execution.IsBid,
		OrderStatus: execution.OrderStatus,
	}
}
