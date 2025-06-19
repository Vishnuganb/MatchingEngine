package orderBook

import (
	"encoding/json"
	"log"

	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	OrderStatusNew         OrderStatus = "0"
	OrderStatusFill        OrderStatus = "2"
	OrderStatusPartialFill OrderStatus = "1"
	OrderStatusCanceled    OrderStatus = "4"
	OrderStatusRejected    OrderStatus = "8"
	OrderStatusPendingNew  OrderStatus = "A"
)

type ExecutionNotifier interface {
	NotifyEventAndTrade(ID string, value json.RawMessage) error
}

type Order struct {
	ID                string          `json:"id"`
	Instrument        string          `json:"instrument"`
	Price             decimal.Decimal `json:"price"`
	OrderQty          decimal.Decimal `json:"order_qty"`
	LeavesQty         decimal.Decimal `json:"leaves_qty"`
	Timestamp         int64           `json:"timestamp"`
	IsBid             bool            `json:"is_bid"`
	OrderStatus       OrderStatus     `json:"order_status"`
	CumQty            decimal.Decimal `json:"exec_qty"`
	ExecutionNotifier ExecutionNotifier
}

func (o *Order) updateOrderQuantities(qty decimal.Decimal) {
	o.LeavesQty = o.LeavesQty.Sub(qty)
	o.CumQty = o.CumQty.Add(qty)
}

func (o *Order) ResetQuantities() {
	o.CumQty = decimal.Zero
	o.LeavesQty = o.OrderQty
}

func (o *Order) SetStatus(status OrderStatus) {
	o.OrderStatus = status
}

// In Order struct methods where you want to publish events
func (o *Order) publishExecutionReport(e Event) {
	if o.ExecutionNotifier == nil {
		log.Printf("Warning: KafkaProducer is nil for order %s", o.ID)
		return
	}

	report := NewExecutionReportFromEvent(e)

	payload, err := json.Marshal(report)
	if err != nil {
		log.Printf("Error marshaling execution report: %v", err)
		return
	}

	if err := o.ExecutionNotifier.NotifyEventAndTrade(e.ID, payload); err != nil {
		log.Printf("Error publishing execution report: %v", err)
	} else {
		log.Printf("Execution report published for event %s [%s]", e.ID, e.OrderStatus)
	}
}
