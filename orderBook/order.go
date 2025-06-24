package orderBook

import (
	"encoding/json"
	"log"

	"github.com/shopspring/decimal"

	"MatchingEngine/internal/model"
	"MatchingEngine/internal/util"
)

type ExecutionNotifier interface {
	NotifyEventAndTrade(ID string, value json.RawMessage) error
}

type Order struct {
	ClOrdID           string            `json:"cl_ord_id"` // from FIX <11>
	OrderID           string            `json:"order_id"`  // from FIX <37>
	Symbol            string            `json:"symbol"`    // from FIX <55>
	Side              model.Side        `json:"side"`      // from FIX <54>
	Price             decimal.Decimal   `json:"price"`     // from FIX <44>`
	OrderQty          decimal.Decimal   `json:"order_qty"` // from FIX <38>
	LeavesQty         decimal.Decimal   `json:"leaves_qty"`
	CumQty            decimal.Decimal   `json:"cum_qty"`
	AvgPx             decimal.Decimal   `json:"avg_px"`
	Timestamp         int64             `json:"transact_time"` // from FIX <60>
	OrderStatus       model.OrderStatus `json:"order_status"`
	Text              string            `json:"text,omitempty"` // from FIX <58>
	ExecutionNotifier ExecutionNotifier
}

func (o *Order) AssignOrderID() {
	o.OrderID = util.GeneratePrefixedID("order")
}

func (o *Order) NewOrderEvent() {
	o.OrderStatus = model.OrderStatusNew
	er := newExecutionReport(o, model.ExecTypeNew)
	er.SetOrdStatus(model.OrderStatusNew)
	o.publishExecutionReport(er)
}

func (o *Order) newFillEvent(price, qty decimal.Decimal) {
	o.CumQty = o.CumQty.Add(qty)
	o.LeavesQty = o.OrderQty.Sub(o.CumQty)

	status := model.OrderStatusFill
	execType := model.ExecTypeFill
	if o.LeavesQty.IsPositive() {
		status = model.OrderStatusPartialFill
	}

	o.OrderStatus = status
	o.AvgPx = computeAvgPx(o.AvgPx, o.CumQty, price, qty)

	er := newExecutionReport(o, execType)
	er.LastShares = qty
	er.LastPx = price
	er.SetOrdStatus(status)
	o.publishExecutionReport(er)
}

func NewFillOrderEvent(order, matchOrder *Order, qty decimal.Decimal) {
	order.newFillEvent(matchOrder.Price, qty)
	matchOrder.newFillEvent(matchOrder.Price, qty)
}

func (o *Order) NewCanceledOrderEvent() {
	log.Printf("Creating canceled event for order: %s", o.OrderID)
	o.OrderStatus = model.OrderStatusCanceled
	o.LeavesQty = decimal.Zero
	er := newExecutionReport(o, model.ExecTypeCanceled)
	er.SetOrdStatus(model.OrderStatusCanceled)
	o.publishExecutionReport(er)
}

func (o *Order) NewCanceledRejectOrderEvent() {
	log.Printf("Creating canceled reject event for order: %s", o.OrderID)
	o.OrderStatus = model.OrderStatusRejected
	er := newExecutionReport(o, model.ExecTypeRejected)
	er.SetOrdStatus(model.OrderStatusRejected)
	o.publishExecutionReport(er)
}

func (o *Order) NewRejectedOrderEvent() {
	log.Printf("Creating rejected event for order: %s", o.OrderID)
	o.OrderStatus = model.OrderStatusRejected
	er := newExecutionReport(o, model.ExecTypeRejected)
	er.ResetQuantities()
	er.SetOrdStatus(model.OrderStatusRejected)
	o.publishExecutionReport(er)
}

func computeAvgPx(currentAvg decimal.Decimal, totalQty, newPx, fillQty decimal.Decimal) decimal.Decimal {
	if totalQty.IsZero() {
		return decimal.Zero
	}
	totalCost := currentAvg.Mul(totalQty.Sub(fillQty)).Add(newPx.Mul(fillQty))
	return totalCost.Div(totalQty)
}

// In Order struct methods where you want to publish events
func (o *Order) publishExecutionReport(er model.ExecutionReport) {
	if o.ExecutionNotifier == nil {
		log.Printf("Warning: KafkaProducer is nil for order %s", o.OrderID)
		return
	}

	payload, err := json.Marshal(er)
	if err != nil {
		log.Printf("Error marshaling execution report: %v", err)
		return
	}

	if err := o.ExecutionNotifier.NotifyEventAndTrade(er.ExecID, payload); err != nil {
		log.Printf("Error publishing execution report: %v", err)
	} else {
		log.Printf("Execution report published for event %s [%s]", er.ExecID, er.OrdStatus)
	}
}
