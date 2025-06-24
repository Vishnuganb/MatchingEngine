package orderBook

import (
	"MatchingEngine/internal/model"
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

func validNewOrderReq(clOrdID string) model.NewOrderRequest {
	return model.NewOrderRequest{
		BaseOrderRequest: model.BaseOrderRequest{
			MsgType:      model.MsgTypeNew,
			ClOrdID:      clOrdID,
			Side:         model.Buy,
			Symbol:       "BTC/USDT",
			TransactTime: time.Now().UnixNano(),
			Text:         "Test order",
		},
		OrderQty: decimal.NewFromInt(10),
		Price:    decimal.NewFromFloat(100.50),
	}
}

func TestNewOrderBook(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	ob, _ := NewOrderBook(mockNotifier)
	assert.NotNil(t, ob)
	assert.Equal(t, 0, ob.Bids.Size())
	assert.Equal(t, 0, ob.Asks.Size())
	assert.Empty(t, ob.orderIndex)
}

func TestOnNewOrder_ValidOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	ob, _ := NewOrderBook(mockNotifier)

	req := validNewOrderReq("CLORD001")
	ob.OnNewOrder(req)

	val, ok := ob.Bids.Get(req.Price)
	assert.True(t, ok)
	orderList := val.(*OrderList)
	assert.Equal(t, req.ClOrdID, orderList.Orders[0].ClOrdID)
	assert.Contains(t, ob.orderIndex, req.ClOrdID)
}

func TestOnNewOrder_InvalidOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	ob, _ := NewOrderBook(mockNotifier)

	req := validNewOrderReq("CLORD002")
	req.Price = decimal.NewFromInt(-100) // Invalid

	ob.OnNewOrder(req)

	assert.Equal(t, 0, ob.Bids.Size())
	assert.NotContains(t, ob.orderIndex, req.ClOrdID)
}

func TestCancelOrder_ExistingOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	ob, _ := NewOrderBook(mockNotifier)

	req := validNewOrderReq("CLORD003")
	ob.OnNewOrder(req)

	ob.CancelOrder("CLORD003")
	assert.Equal(t, 0, ob.Bids.Size())
	assert.NotContains(t, ob.orderIndex, req.ClOrdID)
}

func TestCancelOrder_NonExistingOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	ob, _ := NewOrderBook(mockNotifier)

	ob.CancelOrder("INVALID_ID")
	assert.Equal(t, 0, ob.Bids.Size())
	assert.Equal(t, 0, ob.Asks.Size())
	assert.Empty(t, ob.orderIndex)
}

func TestAddOrderToBook_BuyOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	ob, _ := NewOrderBook(mockNotifier)

	order := Order{
		ClOrdID:     "CLORD004",
		Symbol:      "BTC/USDT",
		Side:        model.Buy,
		Price:       decimal.NewFromInt(120),
		OrderQty:    decimal.NewFromInt(10),
		LeavesQty:   decimal.NewFromInt(10),
		Timestamp:   time.Now().UnixNano(),
		OrderStatus: model.OrderStatusPendingNew,
	}
	ob.addOrderToBook(order)

	val, ok := ob.Bids.Get(order.Price)
	assert.True(t, ok)
	orderList := val.(*OrderList)
	assert.Equal(t, order.ClOrdID, orderList.Orders[0].ClOrdID)
	assert.Contains(t, ob.orderIndex, order.ClOrdID)
}

func TestAddOrderToBook_SellOrder(t *testing.T) {
	mockNotifier := &MockTradeNotifier{}
	ob, _ := NewOrderBook(mockNotifier)

	order := Order{
		ClOrdID:     "CLORD005",
		Symbol:      "BTC/USDT",
		Side:        model.Sell,
		Price:       decimal.NewFromInt(200),
		OrderQty:    decimal.NewFromInt(5),
		LeavesQty:   decimal.NewFromInt(5),
		Timestamp:   time.Now().UnixNano(),
		OrderStatus: model.OrderStatusPendingNew,
	}
	ob.addOrderToBook(order)

	val, ok := ob.Asks.Get(order.Price)
	assert.True(t, ok)
	orderList := val.(*OrderList)
	assert.Equal(t, order.ClOrdID, orderList.Orders[0].ClOrdID)
	assert.Contains(t, ob.orderIndex, order.ClOrdID)
}
