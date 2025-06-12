package orderBook

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"MatchingEngine/internal/model"
)

func TestNewOrderBook(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	book := NewOrderBook(mockProducer)
	assert.NotNil(t, book)
	assert.Empty(t, book.Bids)
	assert.Empty(t, book.Asks)
	assert.Empty(t, book.Orders)
}

func TestAddBuyOrder(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	book := NewOrderBook(mockProducer)
	order := Order{
		ID:         "1",
		Price:      decimal.NewFromInt(100),
		OrderQty:   decimal.NewFromInt(10),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}
	book.AddBuyOrder(order)
	assert.Len(t, book.Bids, 1)
	assert.Equal(t, book.Bids[0].ID, "1")
}

func TestAddSellOrder(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	book := NewOrderBook(mockProducer)
	order := Order{
		ID:         "2",
		Price:      decimal.NewFromInt(200),
		OrderQty:   decimal.NewFromInt(5),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		IsBid:      false,
	}
	book.AddSellOrder(order)
	assert.Len(t, book.Asks, 1)
	assert.Equal(t, book.Asks[0].ID, "2")
}

func TestCancelOrder(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	book := NewOrderBook(mockProducer)
	order := Order{
		ID:         "1",
		Price:      decimal.NewFromInt(100),
		OrderQty:   decimal.NewFromInt(10),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}
	book.AddBuyOrder(order)

	book.CancelOrder("1", mockProducer)
	assert.Len(t, book.Bids, 0)
}

func TestNewOrder(t *testing.T) {
	mockProducer := &MockKafkaProducer{}
	book := NewOrderBook(mockProducer)
	order := model.Order{
		ID:         "1",
		Price:      decimal.NewFromInt(100),
		OrderQty:   decimal.NewFromInt(10),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}
	book.OnNewOrder(order, mockProducer)
	assert.Len(t, book.Bids, 1)
	assert.Equal(t, book.Bids[0].ID, "1")
	assert.Equal(t, book.Bids[0].Price, decimal.NewFromInt(100))
	assert.Equal(t, book.Bids[0].LeavesQty, decimal.NewFromInt(10))
}
