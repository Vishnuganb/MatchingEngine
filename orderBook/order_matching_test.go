package orderBook

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func newTestOrderBook() *OrderBook {
	book := NewOrderBook()
	book.KafkaProducer = &MockKafkaProducer{}
	return book
}


func TestProcessBuyOrder_FullMatch(t *testing.T) {
	book := newTestOrderBook()

	// Add a sell order to the order book
	book.AddSellOrder(Order{
		ID:        "1",
		Price:     decimal.NewFromInt(100),
		Qty:       decimal.NewFromInt(10),
		LeavesQty: decimal.NewFromInt(10),
		Timestamp: time.Now().UnixNano(),
		IsBid:     false,
	})

	// Process a buy order that fully matches the sell order
	buyOrder := Order{
		ID:        "2",
		Price:     decimal.NewFromInt(100),
		Qty:       decimal.NewFromInt(10),
		LeavesQty: decimal.NewFromInt(10),
		Timestamp: time.Now().UnixNano(),
		IsBid:     true,
	}

	trades := book.processBuyOrder(buyOrder)

	assert.Len(t, trades, 1)
	assert.Equal(t, "2", trades[0].BuyerOrderID)
	assert.Equal(t, "1", trades[0].SellerOrderID)
	assert.Equal(t, uint64(10), trades[0].Quantity)
	assert.Equal(t, uint64(100), trades[0].Price)
	assert.Len(t, book.Asks, 0) // Sell order should be removed
}

func TestProcessSellOrder_PartialMatch(t *testing.T) {
	book := newTestOrderBook()

	// Add a buy order to the order book
	book.AddBuyOrder(Order{
		ID:        "1",
		Price:     decimal.NewFromInt(100),
		Qty:       decimal.NewFromInt(10),
		LeavesQty: decimal.NewFromInt(10),
		Timestamp: time.Now().UnixNano(),
		IsBid:     true,
	})

	// Process a sell order that partially matches the buy order
	sellOrder := Order{
		ID:        "2",
		Price:     decimal.NewFromInt(100),
		Qty:       decimal.NewFromInt(5),
		LeavesQty: decimal.NewFromInt(5),
		Timestamp: time.Now().UnixNano(),
		IsBid:     false,
	}

	trades := book.processSellOrder(sellOrder)

	assert.Len(t, trades, 1)
	assert.Equal(t, "1", trades[0].BuyerOrderID)
	assert.Equal(t, "2", trades[0].SellerOrderID)
	assert.Equal(t, uint64(5), trades[0].Quantity)
	assert.Equal(t, uint64(100), trades[0].Price)
	assert.Len(t, book.Bids, 1)                                    // Buy order remains
	assert.Equal(t, decimal.NewFromInt(5), book.Bids[0].LeavesQty) // 5 remaining
}

func TestProcessOrder_NoMatch(t *testing.T) {
	book := newTestOrderBook()

	buyOrder := Order{
		ID:        "1",
		Price:     decimal.NewFromInt(100),
		Qty:       decimal.NewFromInt(10),
		LeavesQty: decimal.NewFromInt(10),
		Timestamp: time.Now().UnixNano(),
		IsBid:     true,
	}

	trades := book.processBuyOrder(buyOrder)

	assert.Len(t, trades, 0)
	assert.Len(t, book.Bids, 1)
	assert.Equal(t, "1", book.Bids[0].ID)
}

func TestProcessOrder_InvalidOrder(t *testing.T) {
	book := newTestOrderBook()

	invalidOrder := Order{
		ID:        "1",
		Price:     decimal.NewFromInt(-100),
		Qty:       decimal.NewFromInt(10),
		LeavesQty: decimal.NewFromInt(10),
		Timestamp: time.Now().UnixNano(),
		IsBid:     true,
	}

	trades := book.processBuyOrder(invalidOrder)

	assert.Len(t, trades, 0)
	assert.Len(t, book.Bids, 0)
	assert.Len(t, book.Events, 1)
	assert.Equal(t, EventTypeRejected, book.Events[0].Type)
	assert.Equal(t, "1", book.Events[0].OrderID)
}
