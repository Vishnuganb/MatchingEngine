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

// In Order struct methods where you want to publish events
func (o *Order) publishExecutionReport(e Event) {
	if o.ExecutionNotifier == nil {
		log.Printf("Warning: KafkaProducer is nil for order %s", o.ID)
		return
	}

	report := ExecutionReport{
		ExecType:    string(e.ExecType),
		OrderID:     e.OrderID,
		Instrument:  e.Instrument,
		Price:       e.Price,
		OrderQty:    e.OrderQty,
		LeavesQty:   e.LeavesQty,
		CumQty:      e.CumQty,
		IsBid:       e.IsBid,
		Timestamp:   e.Timestamp,
		OrderStatus: string(e.OrderStatus),
	}

	payload, err := json.Marshal(report)
	if err != nil {
		log.Printf("Error marshaling execution report: %v", err)
		return
	}

	if err := o.ExecutionNotifier.NotifyEventAndTrade(e.ID, payload); err != nil {
		log.Printf("Error publishing execution report: %v", err)
	} else {
		log.Printf("Execution report published for event %s [%s]", e.ID, e.ExecType)
	}
}
