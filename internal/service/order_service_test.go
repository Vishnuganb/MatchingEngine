package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"MatchingEngine/internal/model"
)

type MockNotifier struct{}

func (m *MockNotifier) NotifyEventAndTrade(orderID string, value json.RawMessage) error {
	return nil
}

func TestProcessOrderRequest_NewOrder(t *testing.T) {
	notifier := &MockNotifier{}
	orderService := NewOrderService(notifier)

	req := model.OrderRequest{
		MsgType: model.MsgTypeNew,
		NewOrderReq: model.NewOrderRequest{
			BaseOrderRequest: model.BaseOrderRequest{
				MsgType:      model.MsgTypeNew,
				ClOrdID:      "CL001",
				Side:         model.Buy,
				Symbol:       "BTC/USDT",
				TransactTime: time.Now().UnixNano(),
			},
			OrderQty: decimal.NewFromInt(10),
			Price:    decimal.NewFromInt(100),
		},
	}

	err := orderService.ProcessOrderRequest(req)
	assert.NoError(t, err)

	orderService.mu.Lock()
	defer orderService.mu.Unlock()

	ch, exists := orderService.orderChannels["BTC/USDT"]
	assert.True(t, exists)

	// ensure message sent to channel
	select {
	case ch <- req:
	default:
		t.Error("Channel is unexpectedly full or blocked")
	}
}

func TestProcessOrderRequest_CancelOrder(t *testing.T) {
	notifier := &MockNotifier{}
	orderService := NewOrderService(notifier)

	req := model.OrderRequest{
		MsgType: model.MsgTypeCancel,
		CancelOrderReq: model.OrderCancelRequest{
			BaseOrderRequest: model.BaseOrderRequest{
				MsgType:      model.MsgTypeCancel,
				ClOrdID:      "CL002",
				Symbol:       "BTC/USDT",
				Side:         model.Buy,
				TransactTime: time.Now().UnixNano(),
			},
			OrigClOrdID: "CL001",
		},
	}

	err := orderService.ProcessOrderRequest(req)
	assert.NoError(t, err)

	orderService.mu.Lock()
	defer orderService.mu.Unlock()

	_, exists := orderService.orderChannels["BTC/USDT"]
	assert.True(t, exists)
}

func TestProcessOrderRequest_InvalidMessageType(t *testing.T) {
	notifier := &MockNotifier{}
	orderService := NewOrderService(notifier)

	req := model.OrderRequest{
		MsgType: "X",
		NewOrderReq: model.NewOrderRequest{
			BaseOrderRequest: model.BaseOrderRequest{
				MsgType:      "X",
				ClOrdID:      "CL001",
				Side:         model.Buy,
				Symbol:       "BTC/USDT",
				TransactTime: time.Now().UnixNano(),
			},
			OrderQty: decimal.NewFromInt(10),
			Price:    decimal.NewFromInt(100),
		},
	}

	err := orderService.ProcessOrderRequest(req)
	assert.Error(t, err)

	orderService.mu.Lock()
	defer orderService.mu.Unlock()

	assert.Equal(t, 0, len(orderService.orderChannels))
}
