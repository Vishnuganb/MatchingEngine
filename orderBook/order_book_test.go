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
	assert.Equal(t, 0, book.Bids.Size())
	assert.Equal(t, 0, book.Asks.Size())
	assert.Empty(t, book.orderIndex)
}

func TestOnNewOrder_ValidOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	book := NewOrderBook(mockNotifier)

	order := OrderRequest{
		ID:         "1",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		Side:       Buy,
	}

	book.OnNewOrder(order)

	val, ok := book.Bids.Get(order.Price)
	assert.True(t, ok)
	orderList := val.(*OrderList)
	assert.Equal(t, "1", orderList.Orders[0].ID)
	assert.Contains(t, book.orderIndex, "1")
}

func TestOnNewOrder_InvalidOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	book := NewOrderBook(mockNotifier)

	order := OrderRequest{
		ID:         "2",
		Price:      decimal.NewFromInt(-100), // Invalid price
		Qty:        decimal.NewFromInt(10),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		Side:       Buy,
	}

	book.OnNewOrder(order)

	assert.Equal(t, 0, book.Bids.Size())
	assert.NotContains(t, book.orderIndex, "2")
}

func TestCancelOrder_ExistingOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	book := NewOrderBook(mockNotifier)

	order := OrderRequest{
		ID:         "1",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		Side:       Buy,
	}

	book.OnNewOrder(order)
	book.CancelOrder("1")

	assert.Equal(t, 0, book.Bids.Size())
	assert.NotContains(t, book.orderIndex, "1")
}

func TestCancelOrder_NonExistingOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	book := NewOrderBook(mockNotifier)

	book.CancelOrder("999") // Non-existing order

	assert.Equal(t, 0, book.Bids.Size())
	assert.Equal(t, 0, book.Asks.Size())
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

	val, ok := book.Bids.Get(order.Price)
	assert.True(t, ok)
	orderList := val.(*OrderList)
	assert.Equal(t, "1", orderList.Orders[0].ID)
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

	val, ok := book.Asks.Get(order.Price)
	assert.True(t, ok)
	orderList := val.(*OrderList)
	assert.Equal(t, "2", orderList.Orders[0].ID)
	assert.Contains(t, book.orderIndex, "2")
}
