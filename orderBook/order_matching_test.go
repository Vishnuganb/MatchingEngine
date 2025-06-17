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
		ExecutionNotifier: book.TradeNotifier, // Assign notifier
	}
	book.processOrder(&buyOrder)

	// Assert that the sell order is removed and the buy order is fully matched
	assert.Len(t, book.Asks, 0) // Sell order should be removed
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
		ExecutionNotifier: book.TradeNotifier, // Assign notifier
	}
	book.processOrder(&sellOrder)

	// Process a buy order that partially matches the sell order
	buyOrder := Order{
		ID:         "2",
		Instrument: "BTC/USDT",
		Price:      decimal.NewFromInt(100),
		OrderQty:   decimal.NewFromInt(5),
		LeavesQty:  decimal.NewFromInt(5),
		Timestamp:  time.Now().UnixNano(),
		OrderStatus:       OrderStatusPendingNew,
		IsBid:      true,
		ExecutionNotifier: book.TradeNotifier, // Assign notifier
	}

	book.processOrder(&buyOrder)

	log.Printf("Seller Order: %+v", sellOrder)

	assert.Len(t, book.Asks, 1)
	assert.True(t, buyOrder.LeavesQty.Equal(decimal.Zero))
	assert.Equal(t, OrderStatusFill, buyOrder.OrderStatus)
	assert.Equal(t, OrderStatusPartialFill, sellOrder.OrderStatus)
}

func TestProcessOrder_NoMatch(t *testing.T) {
	book := newTestOrderBook()

	// Process a buy order with no matching sell orders
	buyOrder := Order{
		ID:         "1",
		Instrument: "BTC/USDT",
		Price:      decimal.NewFromInt(50),
		OrderQty:   decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
		Timestamp:  time.Now().UnixNano(),
		OrderStatus:       OrderStatusPendingNew,
		IsBid:      true,
	}

	book.processOrder(&buyOrder)

	assert.Len(t, book.Bids, 1) // Buy order added to the book
	assert.Equal(t, decimal.NewFromInt(10), book.Bids[decimal.NewFromInt(50)].Orders[0].LeavesQty)
}

func TestProcessOrder_InvalidOrder(t *testing.T) {
	book := newTestOrderBook()

	// Process an invalid order
	invalidOrder := Order{
		ID:         "1",
		Instrument: "BTC/USDT",
		Price:      decimal.NewFromInt(-100), // Invalid price
		OrderQty:   decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
		Timestamp:  time.Now().UnixNano(),
		OrderStatus:       OrderStatusPendingNew,
		IsBid:      true,
	}

	book.processOrder(&invalidOrder)

	assert.Len(t, book.Bids, 0) // Invalid order should not be added
}
