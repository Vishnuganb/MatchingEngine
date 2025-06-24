package orderBook

import (
	"encoding/json"
	"testing"
	"time"

	"MatchingEngine/internal/model"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type MockNotifier struct {
	Published bool
	LastID    string
	LastData  json.RawMessage
}

func (m *MockNotifier) NotifyEventAndTrade(id string, data json.RawMessage) error {
	m.Published = true
	m.LastID = id
	m.LastData = data
	return nil
}

func newTestOrder() *Order {
	return &Order{
		ClOrdID:     "CL123",
		Symbol:      "BTC/USDT",
		Side:        model.Buy,
		Price:       decimal.NewFromInt(100),
		OrderQty:    decimal.NewFromInt(10),
		LeavesQty:   decimal.NewFromInt(10),
		Timestamp:   time.Now().UnixNano(),
		OrderStatus: model.OrderStatusPendingNew,
		CumQty:      decimal.Zero,
		AvgPx:       decimal.Zero,
	}
}

func TestAssignOrderID(t *testing.T) {
	order := newTestOrder()
	order.AssignOrderID()
	assert.NotEmpty(t, order.OrderID)
}

func TestNewOrderEvent(t *testing.T) {
	order := newTestOrder()
	notifier := &MockNotifier{}
	order.ExecutionNotifier = notifier

	order.NewOrderEvent()

	assert.True(t, notifier.Published)
	assert.Equal(t, model.OrderStatusNew, order.OrderStatus)
}

func TestNewRejectedOrderEvent(t *testing.T) {
	order := newTestOrder()
	notifier := &MockNotifier{}
	order.ExecutionNotifier = notifier

	order.NewRejectedOrderEvent()

	assert.True(t, notifier.Published)
	assert.Equal(t, model.OrderStatusRejected, order.OrderStatus)
}

func TestNewCanceledOrderEvent(t *testing.T) {
	order := newTestOrder()
	notifier := &MockNotifier{}
	order.ExecutionNotifier = notifier

	order.NewCanceledOrderEvent()

	assert.True(t, notifier.Published)
	assert.Equal(t, model.OrderStatusCanceled, order.OrderStatus)
	assert.True(t, order.LeavesQty.IsZero())
}

func TestNewCanceledRejectOrderEvent(t *testing.T) {
	order := newTestOrder()
	notifier := &MockNotifier{}
	order.ExecutionNotifier = notifier

	order.NewCanceledRejectOrderEvent()

	assert.True(t, notifier.Published)
	assert.Equal(t, model.OrderStatusRejected, order.OrderStatus)
}

func TestNewFillEvent_FullFill(t *testing.T) {
	order := newTestOrder()
	notifier := &MockNotifier{}
	order.ExecutionNotifier = notifier

	price := decimal.NewFromInt(100)
	qty := decimal.NewFromInt(10)

	order.newFillEvent(price, qty)

	assert.Equal(t, model.OrderStatusFill, order.OrderStatus)
	assert.True(t, order.LeavesQty.IsZero())
	assert.True(t, notifier.Published)
	assert.True(t, order.AvgPx.Equal(price))
}

func TestNewFillEvent_PartialFill(t *testing.T) {
	order := newTestOrder()
	notifier := &MockNotifier{}
	order.ExecutionNotifier = notifier

	price := decimal.NewFromInt(100)
	qty := decimal.NewFromInt(5)

	order.newFillEvent(price, qty)

	assert.Equal(t, model.OrderStatusPartialFill, order.OrderStatus)
	assert.True(t, order.LeavesQty.Equal(decimal.NewFromInt(5)))
	assert.True(t, order.CumQty.Equal(qty))
	assert.True(t, notifier.Published)
}

func TestNewFillOrderEvent(t *testing.T) {
	buy := newTestOrder()
	sell := newTestOrder()
	buy.ExecutionNotifier = &MockNotifier{}
	sell.ExecutionNotifier = &MockNotifier{}

	qty := decimal.NewFromInt(5)
	NewFillOrderEvent(buy, sell, qty)

	assert.Equal(t, decimal.NewFromInt(5), buy.CumQty)
	assert.Equal(t, decimal.NewFromInt(5), sell.CumQty)
}
