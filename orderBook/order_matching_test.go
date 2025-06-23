package orderBook

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"MatchingEngine/internal/model"
)

func newTestOrderBook() *OrderBook {
	mockNotifier := &MockTradeNotifier{}
	book, _ := NewOrderBook(mockNotifier)
	return book
}

func TestProcessOrder_FullMatch(t *testing.T) {
	book := newTestOrderBook()

	sellOrder := Order{
		ClOrdID:           "SELL1",
		OrderID:           "ORDER1",
		Symbol:            "BTC/USDT",
		Side:              model.Sell,
		Price:             decimal.NewFromInt(100),
		OrderQty:          decimal.NewFromInt(10),
		LeavesQty:         decimal.NewFromInt(10),
		Timestamp:         time.Now().UnixNano(),
		OrderStatus:       model.OrderStatusPendingNew,
		ExecutionNotifier: book.TradeNotifier,
	}
	book.addOrderToBook(sellOrder)

	buyOrder := Order{
		ClOrdID:           "BUY1",
		OrderID:           "ORDER2",
		Symbol:            "BTC/USDT",
		Side:              model.Buy,
		Price:             decimal.NewFromInt(100),
		OrderQty:          decimal.NewFromInt(10),
		LeavesQty:         decimal.NewFromInt(10),
		Timestamp:         time.Now().UnixNano(),
		OrderStatus:       model.OrderStatusPendingNew,
		ExecutionNotifier: book.TradeNotifier,
	}
	book.processOrder(&buyOrder)

	assert.Equal(t, 0, book.Asks.Size())
	assert.True(t, buyOrder.LeavesQty.IsZero())
	assert.Equal(t, model.OrderStatusFill, buyOrder.OrderStatus)
}

func TestProcessOrder_PartialMatch(t *testing.T) {
	book := newTestOrderBook()

	sellOrder := Order{
		ClOrdID:           "SELL2",
		OrderID:           "ORDER3",
		Symbol:            "BTC/USDT",
		Side:              model.Sell,
		Price:             decimal.NewFromInt(100),
		OrderQty:          decimal.NewFromInt(10),
		LeavesQty:         decimal.NewFromInt(10),
		Timestamp:         time.Now().UnixNano(),
		OrderStatus:       model.OrderStatusPendingNew,
		ExecutionNotifier: book.TradeNotifier,
	}
	book.addOrderToBook(sellOrder)

	buyOrder := Order{
		ClOrdID:           "BUY2",
		OrderID:           "ORDER4",
		Symbol:            "BTC/USDT",
		Side:              model.Buy,
		Price:             decimal.NewFromInt(100),
		OrderQty:          decimal.NewFromInt(5),
		LeavesQty:         decimal.NewFromInt(5),
		Timestamp:         time.Now().UnixNano(),
		OrderStatus:       model.OrderStatusPendingNew,
		ExecutionNotifier: book.TradeNotifier,
	}
	book.processOrder(&buyOrder)

	assert.Equal(t, 1, book.Asks.Size())
	assert.True(t, buyOrder.LeavesQty.IsZero())
	assert.Equal(t, model.OrderStatusFill, buyOrder.OrderStatus)
}

func TestProcessOrder_NoMatch(t *testing.T) {
	book := newTestOrderBook()

	buyOrder := Order{
		ClOrdID:     "BUY3",
		OrderID:     "ORDER5",
		Symbol:      "BTC/USDT",
		Side:        model.Buy,
		Price:       decimal.NewFromInt(50),
		OrderQty:    decimal.NewFromInt(10),
		LeavesQty:   decimal.NewFromInt(10),
		Timestamp:   time.Now().UnixNano(),
		OrderStatus: model.OrderStatusPendingNew,
	}
	book.processOrder(&buyOrder)

	assert.Equal(t, 1, book.Bids.Size())
	assert.True(t, buyOrder.LeavesQty.Equal(decimal.NewFromInt(10)))
}
