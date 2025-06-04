package orderBook

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestProcessBuyOrder_FullMatch(t *testing.T) {
	book := NewOrderBook()

	// Add a sell order to the order book
	book.AddSellOrder(Order{
		ID:         "1",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
		Timestamp:  time.Now().UnixNano(),
		IsBid:      false,
	})

	// Process a buy order that fully matches the sell order
	buyOrder := Order{
		ID:         "2",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}

	trade := book.processBuyOrder(buyOrder)

	assert.Equal(t, "2", trade.BuyerOrderID)
	assert.Equal(t, "1", trade.SellerOrderID)
	assert.Equal(t, uint64(10), trade.Quantity)
	assert.Equal(t, uint64(100), trade.Price)
	assert.Equal(t, 0, len(book.Asks)) // Sell order should be removed
}

func TestProcessSellOrder_PartialMatch(t *testing.T) {
	book := NewOrderBook()

	// Add a buy order to the order book
	book.AddBuyOrder(Order{
		ID:         "1",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	})

	// Process a sell order that partially matches the buy order
	sellOrder := Order{
		ID:         "2",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(5),
		LeavesQty:  decimal.NewFromInt(5),
		Timestamp:  time.Now().UnixNano(),
		IsBid:      false,
	}

	trade := book.processSellOrder(sellOrder)

	assert.Equal(t, "1", trade.BuyerOrderID)
	assert.Equal(t, "2", trade.SellerOrderID)
	assert.Equal(t, uint64(5), trade.Quantity)
	assert.Equal(t, uint64(100), trade.Price)
	assert.Equal(t, 1, len(book.Bids)) // Buy order should remain
	assert.Equal(t, decimal.NewFromInt(5), book.Bids[0].LeavesQty) // Remaining quantity
}

func TestProcessOrder_NoMatch(t *testing.T) {
	book := NewOrderBook()

	// Process a buy order with no matching sell orders
	buyOrder := Order{
		ID:         "1",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}

	trade := book.processBuyOrder(buyOrder)

	assert.Equal(t, uint64(0), trade.Quantity) // No trade should occur
	assert.Equal(t, 1, len(book.Bids))        // Buy order should be added to the book
	assert.Equal(t, "1", book.Bids[0].ID)
}

func TestProcessOrder_InvalidOrder(t *testing.T) {
	book := NewOrderBook()

	// Process an invalid order (negative price)
	invalidOrder := Order{
		ID:         "1",
		Price:      decimal.NewFromInt(-100),
		Qty:        decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}

	trade := book.processBuyOrder(invalidOrder)

	assert.Equal(t, uint64(0), trade.Quantity) // No trade should occur
	assert.Equal(t, 0, len(book.Bids))        // Order should not be added to the book
	assert.Equal(t, 1, len(book.Events))      // Rejected event should be created
	assert.Equal(t, EventTypeRejected, book.Events[0].Type)
}