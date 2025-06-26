package handler

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/mock"

	"MatchingEngine/internal/model"
)

type mockAcknowledger struct{}

func (m *mockAcknowledger) Ack(tag uint64, multiple bool) error {
	return nil
}

func (m *mockAcknowledger) Nack(tag uint64, multiple, requeue bool) error {
	return nil
}

func (m *mockAcknowledger) Reject(tag uint64, requeue bool) error {
	return nil
}

type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) ProcessOrderRequest(req model.OrderRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func TestHandleOrderMessage_ValidOrder(t *testing.T) {
	mockOrderService := new(MockOrderService)
	h := NewOrderRequestHandler(mockOrderService)

	orderReq := model.OrderRequest{
		MsgType: model.MsgTypeNew,
		NewOrderReq: model.NewOrderRequest{
			BaseOrderRequest: model.BaseOrderRequest{
				ClOrdID:      "cl123",
				Side:         model.Buy,
				Symbol:       "BTC/USDT",
				TransactTime: time.Now().UnixNano(),
			},
			OrderQty: decimal.NewFromInt(10),
			Price:    decimal.NewFromInt(100),
		},
	}

	mockOrderService.On("ProcessOrderRequest", orderReq).Return(nil)

	body, _ := json.Marshal(orderReq)
	msg := amqp.Delivery{Body: body, Acknowledger: &mockAcknowledger{}}

	h.HandleOrderMessage(msg)
	mockOrderService.AssertExpectations(t)
}

func TestHandleOrderMessage_InvalidJSON(t *testing.T) {
	mockOrderService := new(MockOrderService)
	h := NewOrderRequestHandler( mockOrderService)

	msg := amqp.Delivery{
		Body:         []byte("{invalid json"),
		Acknowledger: &mockAcknowledger{},
	}

	h.HandleOrderMessage(msg)
}
