package orderBook

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewOrderBook(t *testing.T) {
	book := NewOrderBook()
	assert.NotNil(t, book)
	assert.Empty(t, book.Bids)
	assert.Empty(t, book.Asks)
	assert.Empty(t, book.Events)
}

func TestAddBuyOrder(t *testing.T) {
	book := NewOrderBook()
	order := Order{
		ID:         "1",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}
	book.AddBuyOrder(order)
	assert.Len(t, book.Bids, 1)
	assert.Equal(t, book.Bids[0].ID, "1")
}

func TestAddSellOrder(t *testing.T) {
	book := NewOrderBook()
	order := Order{
		ID:         "2",
		Price:      decimal.NewFromInt(200),
		Qty:        decimal.NewFromInt(5),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		IsBid:      false,
	}
	book.AddSellOrder(order)
	assert.Len(t, book.Asks, 1)
	assert.Equal(t, book.Asks[0].ID, "2")
}

func TestProcessBuyOrder(t *testing.T) {
	book := NewOrderBook()
	sellOrder := Order{
		ID:         "2",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(5),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		IsBid:      false,
	}
	book.AddSellOrder(sellOrder)

	buyOrder := Order{
		ID:         "1",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(5),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}
	trade := book.processBuyOrder(buyOrder)

	assert.Equal(t, uint64(5), trade.Quantity)
	assert.Equal(t, uint64(100), trade.Price)
	assert.Empty(t, book.Asks)
}

func TestCancelOrder(t *testing.T) {
	book := NewOrderBook()
	order := Order{
		ID:         "1",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}
	book.AddBuyOrder(order)

	event := book.CancelOrder("1")
	assert.Equal(t, EventTypeCanceled, event.Type)
	assert.Empty(t, book.Bids)
}

func TestNewOrder(t *testing.T) {
	book := NewOrderBook()
	order := Order{
		ID:         "1",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}
	event := book.NewOrder(order)
	assert.Equal(t, EventTypeNew, event.Type)
	assert.Len(t, book.Bids, 1)
}