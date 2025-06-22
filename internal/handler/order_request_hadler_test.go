package handler

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"MatchingEngine/internal/model"
	"MatchingEngine/internal/rmq"
	"MatchingEngine/orderBook"
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

func (m *MockTradeService) SaveTradeAsync(trade model.Trade) {
	m.Called(trade)
}

type mockAcknowledger struct{}

func (m *mockAcknowledger) Ack(uint64, bool) error {
	return nil
}

func (m *mockAcknowledger) Nack(uint64, bool, bool) error {
	return nil
}

func (m *mockAcknowledger) Reject(uint64, bool) error {
	return nil
}

type MockTradeNotifier struct{}

func (m *MockTradeNotifier) NotifyEventAndTrade(string, json.RawMessage) error {
	return nil
}

func TestHandleOrderMessage_ValidOrder(t *testing.T) {
	mockService := new(MockExecutionService)
	mockTradeService := new(MockTradeService)
	mockNotifier := new(MockTradeNotifier)
	handler := NewOrderRequestHandler(mockService, mockTradeService, mockNotifier)

	orderReq := model.OrderRequest{
		RequestType: rmq.ReqTypeNew,
		Order: rmq.TraderOrder{
			ID:         "1",
			Price:      "100",
			Qty:        "10",
			Instrument: "BTC/USDT",
			Side:       orderBook.Buy,
		},
	}

	body, _ := json.Marshal(orderReq)
	msg := amqp.Delivery{Body: body, Acknowledger: &mockAcknowledger{}}

	handler.HandleOrderMessage(msg)

	handler.mu.Lock()
	orderChannel, exists := handler.orderChannels["BTC/USDT"]
	handler.mu.Unlock()

	assert.True(t, exists)
	assert.NotNil(t, orderChannel)
}

func TestHandleOrderMessage_InvalidJSON(t *testing.T) {
	mockService := new(MockExecutionService)
	mockTradeService := new(MockTradeService)
	mockNotifier := new(MockTradeNotifier)
	handler := NewOrderRequestHandler(mockService, mockTradeService, mockNotifier)

	msg := amqp.Delivery{
		Body:         []byte("invalid json"),
		Acknowledger: &mockAcknowledger{},
	}

	handler.HandleOrderMessage(msg)

	handler.mu.Lock()
	assert.Equal(t, 0, len(handler.orderChannels), "No order channels should be created for invalid JSON")
	handler.mu.Unlock()
}

func TestStartOrderWorkerForInstrument(t *testing.T) {
	mockService := new(MockExecutionService)
	mockTradeService := new(MockTradeService)
	mockNotifier := new(MockTradeNotifier)
	handler := NewOrderRequestHandler(mockService, mockTradeService, mockNotifier)

	instrument := "BTC/USDT"
	orderReq := model.OrderRequest{
		RequestType: rmq.ReqTypeNew,
		Order: rmq.TraderOrder{
			ID:         "1",
			Price:      "100",
			Qty:        "10",
			Instrument: instrument,
			Side:       orderBook.Buy,
		},
	}

	channel := make(chan model.OrderRequest, 100)
	handler.mu.Lock()
	handler.orderChannels[instrument] = channel
	handler.mu.Unlock()

	go handler.startOrderWorkerForInstrument(instrument, channel)

	channel <- orderReq

	time.Sleep(100 * time.Millisecond)

	handler.mu.Lock()
	book, exists := handler.orderBooks[instrument]
	handler.mu.Unlock()

	assert.True(t, exists)
	assert.NotNil(t, book)
}
