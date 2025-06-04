package orderBook

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewRejectedEvent(t *testing.T) {
	orderRequest := &OrderRequest{
		ID:        "1",
		Side:      Buy,
		Price:     decimal.NewFromInt(100),
		Qty:       decimal.NewFromInt(10),
		Timestamp: time.Now().UnixNano(),
	}

	event := newRejectedEvent(orderRequest)

	assert.Equal(t, "1", event.OrderID, "OrderID should match")
	assert.Equal(t, EventTypeRejected, event.Type, "Event type should be 'rejected'")
	assert.Equal(t, Buy, event.Side, "Side should match the order request")
	assert.Equal(t, decimal.NewFromInt(10), event.OrderQty, "OrderQty should match the order request")
	assert.Equal(t, decimal.Zero, event.LeavesQty, "LeavesQty should be zero for rejected events")
}

func TestNewFillEvent(t *testing.T) {
	order := &Order{
		ID:         "2",
		Instrument: "BTC/USDT",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(5),
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}

	execQty := decimal.NewFromInt(5)
	price := decimal.NewFromInt(100)

	event := newFillEvent(order, execQty, price)

	assert.Equal(t, "2", event.OrderID, "OrderID should match")
	assert.Equal(t, EventTypePartialFill, event.Type, "Event type should be 'partially fill'")
	assert.Equal(t, Buy, event.Side, "Side should match the order")
	assert.Equal(t, decimal.NewFromInt(10), event.OrderQty, "OrderQty should match the order")
	assert.Equal(t, decimal.NewFromInt(5), event.LeavesQty, "LeavesQty should match the remaining quantity")
	assert.Equal(t, execQty, event.ExecQty, "ExecQty should match the executed quantity")
	assert.Equal(t, price, event.Price, "Price should match the executed price")
}

func TestNewCanceledEvent(t *testing.T) {
	order := &Order{
		ID:         "3",
		Instrument: "BTC/USDT",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
		Timestamp:  time.Now().UnixNano(),
		IsBid:      false,
	}

	event := newCanceledEvent(order)

	assert.Equal(t, "3", event.OrderID, "OrderID should match")
	assert.Equal(t, EventTypeCanceled, event.Type, "Event type should be 'canceled'")
	assert.Equal(t, Sell, event.Side, "Side should match the order")
	assert.Equal(t, decimal.NewFromInt(10), event.OrderQty, "OrderQty should match the order")
	assert.Equal(t, decimal.Zero, event.LeavesQty, "LeavesQty should be zero for canceled events")
}