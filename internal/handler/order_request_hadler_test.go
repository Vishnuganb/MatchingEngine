package handler

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
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
	mockExecSvc := new(MockExecutionService)
	mockTradeSvc := new(MockTradeService)
	mockOrderSvc := new(MockOrderService)

	handler := NewOrderRequestHandler(mockExecSvc, mockTradeSvc, mockOrderSvc)

	execReportJSON := `{
		"35": "8",
		"17": "exec123",
		"37": "order123",
		"11": "cl123",
		"150": "0",
		"39": "2",
		"55": "BTC/USDT",
		"54": "1",
		"38": "10",
		"32": "10",
		"31": "100",
		"151": "0",
		"14": "10",
		"6": "100",
		"60": 1620000000
	}`

	expected := model.ExecutionReport{
		MsgType:      "8",
		ExecID:       "exec123",
		OrderID:      "order123",
		ClOrdID:      "cl123",
		ExecType:     "0",
		OrdStatus:    "2",
		Symbol:       "BTC/USDT",
		Side:         "1",
		OrderQty:     decimal.NewFromInt(10),
		LastShares:   decimal.NewFromInt(10),
		LastPx:       decimal.NewFromInt(100),
		LeavesQty:    decimal.NewFromInt(0),
		CumQty:       decimal.NewFromInt(10),
		AvgPx:        decimal.NewFromInt(100),
		TransactTime: 1620000000,
	}

	mockExecSvc.On("SaveExecutionAsync", mock.MatchedBy(func(r model.ExecutionReport) bool {
		return r.ExecID == expected.ExecID &&
			r.OrderID == expected.OrderID &&
			r.CumQty.Equal(expected.CumQty) &&
			r.LastPx.Equal(expected.LastPx)
	})).Once()

	err := handler.HandleExecutionReport([]byte(execReportJSON))
	assert.NoError(t, err)

	mockExecSvc.AssertExpectations(t)
}

func TestHandleExecutionReport_TradeReport(t *testing.T) {
	mockExecService := new(MockExecutionService)
	mockTradeService := new(MockTradeService)
	handler := NewOrderRequestHandler(mockExecService, mockTradeService, nil)

	rawJSON := []byte(`{
		"35": "AE",
		"571": "trade123",
		"17": "exec123",
		"55": "BTC/USDT",
		"32": "5",
		"31": "100",
		"75": "20250624",
		"60": 1729811234567890,
		"552": [
			{"54": "1", "37": "order1"},
			{"54": "2", "37": "order2"}
		]
	}`)

	expectedReport := model.TradeCaptureReport{
		MsgType:       "AE",
		TradeReportID: "trade123",
		ExecID:        "exec123",
		Symbol:        "BTC/USDT",
		LastQty:       decimal.NewFromInt(5),
		LastPx:        decimal.NewFromInt(100),
		TradeDate:     "20250624",
		TransactTime:  1729811234567890,
		NoSides: []model.NoSides{
			{Side: model.Buy, OrderID: "order1"},
			{Side: model.Sell, OrderID: "order2"},
		},
	}

	mockTradeService.On("SaveTradeAsync", expectedReport).Once()

	err := handler.HandleExecutionReport(rawJSON)
	assert.NoError(t, err)

	mockTradeService.AssertExpectations(t)
}

func TestHandleExecutionReport_InvalidJSON(t *testing.T) {
	mockExecSvc := new(MockExecutionService)
	mockTradeSvc := new(MockTradeService)
	mockOrderSvc := new(MockOrderService)

	handler := NewOrderRequestHandler(mockExecSvc, mockTradeSvc, mockOrderSvc)

	invalidJSON := []byte(`not a valid json`)
	err := handler.HandleExecutionReport(invalidJSON)
	assert.NoError(t, err)
}

func TestHandleExecutionReport_UnknownMsgType(t *testing.T) {
	mockExecSvc := new(MockExecutionService)
	mockTradeSvc := new(MockTradeService)
	mockOrderSvc := new(MockOrderService)

	handler := NewOrderRequestHandler(mockExecSvc, mockTradeSvc, mockOrderSvc)

	unknownMsg := `{"35": "ZZ", "some": "field"}`
	err := handler.HandleExecutionReport([]byte(unknownMsg))
	assert.NoError(t, err)
}
