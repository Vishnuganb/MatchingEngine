package orderBook

import (
	"time"

	"github.com/shopspring/decimal"

	"MatchingEngine/internal/model"
	"MatchingEngine/internal/util"
)

func newExecutionReport(order *Order, execType model.ExecType) model.ExecutionReport {
	return model.ExecutionReport{
		MsgType:      "8",
		ExecID:       util.GeneratePrefixedID("execution"),
		OrderID:      order.OrderID,
		ClOrdID:      order.ClOrdID,
		ExecType:     execType,
		OrdStatus:    order.OrderStatus,
		Symbol:       order.Symbol,
		Side:         order.Side,
		OrderQty:     order.OrderQty,
		LastShares:   decimal.Zero, // updated later if it's a fill
		LastPx:       order.Price,  // updated later if it's a fill
		LeavesQty:    order.LeavesQty,
		CumQty:       order.CumQty,
		AvgPx:        order.AvgPx,
		TransactTime: time.Now().UnixNano(),
	}
}
