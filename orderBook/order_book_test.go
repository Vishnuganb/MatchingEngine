package orderBook

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type MockTradeNotifier struct{}

func (m *MockTradeNotifier) NotifyEventAndTrade(string, json.RawMessage) error {
	return nil
}

func TestNewOrderBook(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	book := NewOrderBook(mockNotifier)
	assert.NotNil(t, book)
	assert.Empty(t, book.Bids)
	assert.Empty(t, book.Asks)
	assert.Empty(t, book.orderIndex)
}

func TestOnNewOrder_ValidOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	book := NewOrderBook(mockNotifier)

	order := Order{
		ID:          "1",
		Price:       decimal.NewFromInt(100),
		OrderQty:    decimal.NewFromInt(10),
		Instrument:  "BTC/USDT",
		Timestamp:   time.Now().UnixNano(),
		OrderStatus: OrderStatusPendingNew,
		IsBid:       true,
	}

	book.OnNewOrder(order)

	assert.Len(t, book.Bids, 1)
	assert.Contains(t, book.Bids, order.Price)
	assert.Equal(t, book.Bids[order.Price].Orders[0].ID, "1")
	assert.Contains(t, book.orderIndex, "1")
}

func TestOnNewOrder_InvalidOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	book := NewOrderBook(mockNotifier)

	order := Order{
		ID:          "2",
		Price:       decimal.NewFromInt(-100), // Invalid price
		OrderQty:    decimal.NewFromInt(10),
		Instrument:  "BTC/USDT",
		Timestamp:   time.Now().UnixNano(),
		OrderStatus: OrderStatusPendingNew,
		IsBid:       true,
	}

	book.OnNewOrder(order)

	assert.Empty(t, book.Bids)
	assert.NotContains(t, book.orderIndex, "2")
}

func TestCancelOrder_ExistingOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	book := NewOrderBook(mockNotifier)

	order := Order{
		ID:          "1",
		Price:       decimal.NewFromInt(100),
		OrderQty:    decimal.NewFromInt(10),
		Instrument:  "BTC/USDT",
		Timestamp:   time.Now().UnixNano(),
		OrderStatus: OrderStatusPendingNew,
		IsBid:       true,
	}

	book.OnNewOrder(order)
	book.CancelOrder("1")

	assert.Empty(t, book.Bids)
	assert.NotContains(t, book.orderIndex, "1")
}

func TestCancelOrder_NonExistingOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	book := NewOrderBook(mockNotifier)

	book.CancelOrder("999") // Non-existing order

	assert.Empty(t, book.Bids)
	assert.Empty(t, book.Asks)
	assert.Empty(t, book.orderIndex)
}

func TestAddOrderToBook_BuyOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	book := NewOrderBook(mockNotifier)

	order := Order{
		ID:          "1",
		Price:       decimal.NewFromInt(100),
		OrderQty:    decimal.NewFromInt(10),
		Instrument:  "BTC/USDT",
		Timestamp:   time.Now().UnixNano(),
		OrderStatus: OrderStatusPendingNew,
		IsBid:       true,
	}

	book.addOrderToBook(order)

	assert.Len(t, book.Bids, 1)
	assert.Contains(t, book.Bids, order.Price)
	assert.Equal(t, book.Bids[order.Price].Orders[0].ID, "1")
	assert.Contains(t, book.orderIndex, "1")
}

func TestAddOrderToBook_SellOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	book := NewOrderBook(mockNotifier)

	order := Order{
		ID:          "2",
		Price:       decimal.NewFromInt(200),
		OrderQty:    decimal.NewFromInt(5),
		Instrument:  "BTC/USDT",
		Timestamp:   time.Now().UnixNano(),
		OrderStatus: OrderStatusPendingNew,
		IsBid:       false,
	}

	book.addOrderToBook(order)

	assert.Len(t, book.Asks, 1)
	assert.Contains(t, book.Asks, order.Price)
	assert.Equal(t, book.Asks[order.Price].Orders[0].ID, "2")
	assert.Contains(t, book.orderIndex, "2")
}
