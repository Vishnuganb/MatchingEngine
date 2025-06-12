package orderBook

import (
	"log"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"MatchingEngine/internal/model"
)

func TestNewOrderBook(t *testing.T) {
	book := NewOrderBook()
	assert.NotNil(t, book)
	assert.Empty(t, book.Bids)
	assert.Empty(t, book.Asks)
	assert.Empty(t, book.Orders)
}

func TestAddBuyOrder(t *testing.T) {
	book := NewOrderBook()
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
	book := NewOrderBook()
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
	book := NewOrderBook()
	order := Order{
		ID:         "1",
		Price:      decimal.NewFromInt(100),
		OrderQty:   decimal.NewFromInt(10),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}
	book.AddBuyOrder(order)

	event := book.CancelOrder("1")
	assert.Equal(t, "canceled", event.ExecType)
	assert.Empty(t, book.Bids)
}

func TestNewOrder(t *testing.T) {
	book := NewOrderBook()
	order := model.Order{
		ID:         "1",
		Price:      decimal.NewFromInt(100),
		OrderQty:   decimal.NewFromInt(10),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}
	orders := book.OnNewOrder(order)
	log.Println(orders)
	for _, order = range orders {
		assert.Equal(t, "new", order.ExecType)
		assert.Equal(t, order.Instrument, order.Instrument)
		assert.Equal(t, order.Price, order.Price)
		assert.Equal(t, order.OrderQty, order.OrderQty)
	}
	assert.Len(t, book.Bids, 1)
}
