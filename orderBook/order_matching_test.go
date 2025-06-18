package orderBook

import (
	"log"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func newTestOrderBook() *OrderBook {
	mockNotifier := &MockTradeNotifier{}
	return NewOrderBook(mockNotifier)
}

func TestProcessOrder_FullMatch(t *testing.T) {
	book := newTestOrderBook()

	// Add a sell order to the order book
	sellOrder := Order{
		ID:                "1",
		Instrument:        "BTC/USDT",
		Price:             decimal.NewFromInt(100),
		OrderQty:          decimal.NewFromInt(10),
		LeavesQty:         decimal.NewFromInt(10),
		Timestamp:         time.Now().UnixNano(),
		OrderStatus:       OrderStatusPendingNew,
		IsBid:             false,
		ExecutionNotifier: book.TradeNotifier, // Assign notifier
	}
	book.addOrderToBook(sellOrder)

	// Process a buy order that fully matches the sell order
	buyOrder := Order{
		ID:                "2",
		Instrument:        "BTC/USDT",
		Price:             decimal.NewFromInt(100),
		OrderQty:          decimal.NewFromInt(10),
		LeavesQty:         decimal.NewFromInt(10),
		Timestamp:         time.Now().UnixNano(),
		OrderStatus:       OrderStatusPendingNew,
		IsBid:             true,
		ExecutionNotifier: book.TradeNotifier,
	}
	book.processOrder(&buyOrder)

	// Assert that the sell order is removed and the buy order is fully matched
	assert.Equal(t, 0, book.Asks.Size()) // Sell order should be removed
	assert.True(t, buyOrder.LeavesQty.Equal(decimal.Zero))
	assert.Equal(t, OrderStatusFill, buyOrder.OrderStatus)
}

func TestProcessOrder_PartialMatch(t *testing.T) {
	book := newTestOrderBook()

	// Add a sell order to the order book
	sellOrder := Order{
		ID:                "1",
		Instrument:        "BTC/USDT",
		Price:             decimal.NewFromInt(100),
		OrderQty:          decimal.NewFromInt(10),
		LeavesQty:         decimal.NewFromInt(10),
		Timestamp:         time.Now().UnixNano(),
		OrderStatus:       OrderStatusPendingNew,
		IsBid:             false,
		ExecutionNotifier: book.TradeNotifier,
	}
	book.addOrderToBook(sellOrder)

	// Process a buy order that partially matches the sell order
	buyOrder := Order{
		ID:                "2",
		Instrument:        "BTC/USDT",
		Price:             decimal.NewFromInt(100),
		OrderQty:          decimal.NewFromInt(5),
		LeavesQty:         decimal.NewFromInt(5),
		Timestamp:         time.Now().UnixNano(),
		OrderStatus:       OrderStatusPendingNew,
		IsBid:             true,
		ExecutionNotifier: book.TradeNotifier,
	}

	book.processOrder(&buyOrder)

	log.Printf("Seller Order: %+v", sellOrder)

	assert.Equal(t, 1, book.Asks.Size())
	assert.True(t, buyOrder.LeavesQty.Equal(decimal.Zero))
	assert.Equal(t, OrderStatusFill, buyOrder.OrderStatus)
}

func TestProcessOrder_NoMatch(t *testing.T) {
	book := newTestOrderBook()

	// Process a buy order with no matching sell orders
	buyOrder := Order{
		ID:          "1",
		Instrument:  "BTC/USDT",
		Price:       decimal.NewFromInt(50),
		OrderQty:    decimal.NewFromInt(10),
		LeavesQty:   decimal.NewFromInt(10),
		Timestamp:   time.Now().UnixNano(),
		OrderStatus: OrderStatusPendingNew,
		IsBid:       true,
	}

	book.processOrder(&buyOrder)

	assert.Equal(t, 1, book.Bids.Size()) // Buy order added to the book
	assert.True(t, buyOrder.LeavesQty.Equal(decimal.NewFromInt(10)))
}
