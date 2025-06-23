package service

import (
	"testing"

	"MatchingEngine/internal/model"
	"MatchingEngine/orderBook"
	"github.com/stretchr/testify/assert"
)

type MockTradeNotifier struct{}

func (m *MockTradeNotifier) NotifyEventAndTrade(orderID string, value []byte) error {
	return nil
}

func TestProcessOrderRequest_NewOrder(t *testing.T) {
	tradeNotifier := &MockTradeNotifier{}
	orderService := NewOrderService(tradeNotifier)

	req := model.OrderRequest{
		MsgType: model.MsgTypeNew,
		NewOrderReq: model.NewOrderRequest{
			Symbol: "BTC/USDT",
			Side:   model.Buy,
		},
	}

	err := orderService.ProcessOrderRequest(req)
	assert.NoError(t, err)

	orderService.mu.Lock()
	defer orderService.mu.Unlock()

	_, exists := orderService.orderBooks["BTC/USDT"]
	assert.True(t, exists)

	_, exists = orderService.orderChannels["BTC/USDT"]
	assert.True(t, exists)
}

func TestProcessOrderRequest_CancelOrder(t *testing.T) {
	tradeNotifier := &MockTradeNotifier{}
	orderService := NewOrderService(tradeNotifier)

	req := model.OrderRequest{
		MsgType: model.MsgTypeCancel,
		CancelOrderReq: model.OrderCancelRequest{
			Symbol: "BTC/USDT",
		},
	}

	err := orderService.ProcessOrderRequest(req)
	assert.NoError(t, err)

	orderService.mu.Lock()
	defer orderService.mu.Unlock()

	_, exists := orderService.orderBooks["BTC/USDT"]
	assert.True(t, exists)

	_, exists = orderService.orderChannels["BTC/USDT"]
	assert.True(t, exists)
}

func TestProcessOrderRequest_InvalidMessageType(t *testing.T) {
	tradeNotifier := &MockTradeNotifier{}
	orderService := NewOrderService(tradeNotifier)

	req := model.OrderRequest{
		MsgType: "InvalidType",
	}

	err := orderService.ProcessOrderRequest(req)
	assert.NoError(t, err)

	orderService.mu.Lock()
	defer orderService.mu.Unlock()

	assert.Equal(t, 0, len(orderService.orderBooks))
	assert.Equal(t, 0, len(orderService.orderChannels))
}
