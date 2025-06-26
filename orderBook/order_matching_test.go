package orderBook

import (
	"testing"
	"time"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"MatchingEngine/internal/model"
	"MatchingEngine/internal/util"
)

func newTestOrderBook() *OrderBook {
	return &OrderBook{
		Bids:       treemap.NewWith(util.DecimalDescComparator),
		Asks:       treemap.NewWith(util.DecimalAscComparator),
		Notifier:   &MockNotifier{},
		orderIndex: make(map[string]*OrderRef),
	}
}

func TestProcessOrder_FullMatch(t *testing.T) {
	book := newTestOrderBook()

	sellOrder := Order{
		ClOrdID:     "CLORD001",
		OrderID:     "ORDER001",
		Symbol:      "BTC/USDT",
		Side:        model.Sell,
		Price:       decimal.NewFromInt(100),
		OrderQty:    decimal.NewFromInt(10),
		LeavesQty:   decimal.NewFromInt(10),
		Timestamp:   time.Now().UnixNano(),
		OrderStatus: model.OrderStatusPendingNew,
	}
	book.addOrderToBook(sellOrder)

	buyOrder := Order{
		ClOrdID:     "CLORD002",
		OrderID:     "ORDER002",
		Symbol:      "BTC/USDT",
		Side:        model.Buy,
		Price:       decimal.NewFromInt(100),
		OrderQty:    decimal.NewFromInt(10),
		LeavesQty:   decimal.NewFromInt(10),
		Timestamp:   time.Now().UnixNano(),
		OrderStatus: model.OrderStatusPendingNew,
	}
	book.processOrder(&buyOrder)

	assert.Equal(t, 0, book.Asks.Size())
	assert.True(t, buyOrder.LeavesQty.IsZero())
}

func TestProcessOrder_PartialMatch(t *testing.T) {
	book := newTestOrderBook()

	sellOrder := Order{
		ClOrdID:     "CLORD001",
		OrderID:     "ORDER001",
		Symbol:      "BTC/USDT",
		Side:        model.Sell,
		Price:       decimal.NewFromInt(100),
		OrderQty:    decimal.NewFromInt(10),
		LeavesQty:   decimal.NewFromInt(10),
		Timestamp:   time.Now().UnixNano(),
		OrderStatus: model.OrderStatusPendingNew,
	}
	book.addOrderToBook(sellOrder)

	buyOrder := Order{
		ClOrdID:     "CLORD002",
		OrderID:     "ORDER002",
		Symbol:      "BTC/USDT",
		Side:        model.Buy,
		Price:       decimal.NewFromInt(100),
		OrderQty:    decimal.NewFromInt(5),
		LeavesQty:   decimal.NewFromInt(5),
		Timestamp:   time.Now().UnixNano(),
		OrderStatus: model.OrderStatusPendingNew,
	}
	book.processOrder(&buyOrder)

	assert.Equal(t, 1, book.Asks.Size())
	assert.True(t, buyOrder.LeavesQty.IsZero())
}

func TestProcessOrder_NoMatch(t *testing.T) {
	book := newTestOrderBook()

	buyOrder := Order{
		ClOrdID:     "CLORD001",
		OrderID:     "ORDER001",
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
