package handler

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"testing"
	"time"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"MatchingEngine/internal/model"
)

type MockExecutionService struct {
	mock.Mock
}

func (m *MockExecutionService) SaveExecutionAsync(order model.ExecutionReport) {
	m.Called(order)
}

type MockTradeService struct {
	mock.Mock
}

func (m *MockTradeService) SaveTradeAsync(trade model.TradeCaptureReport) {
	m.Called(trade)
}

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
	mockExecService := new(MockExecutionService)
	mockTradeService := new(MockTradeService)
	h := NewOrderRequestHandler(mockExecService, mockTradeService, mockOrderService)

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
	mockExecService := new(MockExecutionService)
	mockTradeService := new(MockTradeService)
	h := NewOrderRequestHandler(mockExecService, mockTradeService, mockOrderService)

	msg := amqp.Delivery{
		Body:         []byte("{invalid json"),
		Acknowledger: &mockAcknowledger{},
	}

	h.HandleOrderMessage(msg)
}

func TestHandleExecutionReport_ExecReport(t *testing.T) {
	exec := model.ExecutionReport{
		MsgType: string(model.MsgTypeExecRpt),
		ExecID:  "exec123",
		ClOrdID: "cl123",
	}
	mockOrderService := new(MockOrderService)
	mockExecService := new(MockExecutionService)
	mockTradeService := new(MockTradeService)

	h := NewOrderRequestHandler(mockExecService, mockTradeService, mockOrderService)
	mockExecService.On("SaveExecutionAsync", exec).Return()

	data, _ := json.Marshal(exec)
	h.HandleExecutionReport(data)
	mockExecService.AssertExpectations(t)
}

func TestHandleExecutionReport_TradeReport(t *testing.T) {
	trade := model.TradeCaptureReport{
		MsgType:       string(model.MsgTypeTradeReport),
		TradeReportID: "trade123",
		ExecID:        "exec123",
	}
	mockOrderService := new(MockOrderService)
	mockExecService := new(MockExecutionService)
	mockTradeService := new(MockTradeService)

	h := NewOrderRequestHandler(mockExecService, mockTradeService, mockOrderService)
	mockTradeService.On("SaveTradeAsync", trade).Return()

	data, _ := json.Marshal(trade)
	h.HandleExecutionReport(data)
	mockTradeService.AssertExpectations(t)
}

func TestHandleExecutionReport_InvalidJSON(t *testing.T) {
	mockOrderService := new(MockOrderService)
	mockExecService := new(MockExecutionService)
	mockTradeService := new(MockTradeService)

	h := NewOrderRequestHandler(mockExecService, mockTradeService, mockOrderService)
	invalid := []byte("not-json")
	err := h.HandleExecutionReport(invalid)
	assert.NoError(t, err)
}

func TestHandleExecutionReport_UnknownType(t *testing.T) {
	msg := map[string]interface{}{
		"MsgType": "XYZ",
	}
	data, _ := json.Marshal(msg)

	mockOrderService := new(MockOrderService)
	mockExecService := new(MockExecutionService)
	mockTradeService := new(MockTradeService)

	h := NewOrderRequestHandler(mockExecService, mockTradeService, mockOrderService)
	err := h.HandleExecutionReport(data)
	assert.NoError(t, err)
}
