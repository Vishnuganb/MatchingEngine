package handler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"MatchingEngine/internal/model"
	"MatchingEngine/internal/rmq"
	"MatchingEngine/orderBook"
)

type MockOrderService struct {
	mock.Mock
}

func (m *MockOrderService) SaveOrderAsync(order model.Order) {
	m.Called(order)
}

func (m *MockOrderService) UpdateOrderAsync(orderID, orderStatus, execType string, leavesQty, execQty decimal.Decimal) {
	m.Called(orderID, orderStatus, execType, leavesQty, execQty)
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

type MockEventNotifier struct{}

func (m *MockEventNotifier) NotifyEventAndTrade(orderID string, value json.RawMessage) error {
	return nil
}

func TestHandleEventMessages(t *testing.T) {
	tests := []struct {
		name         string
		eventType    string
		expectSave   bool
		expectUpdate bool
	}{
		{
			name:         "New Order Event",
			eventType:    string(orderBook.EventTypeNew),
			expectSave:   true,
			expectUpdate: false,
		},
		{
			name:         "Pending New Order Event",
			eventType:    string(orderBook.EventTypePendingNew),
			expectSave:   true,
			expectUpdate: false,
		},
		{
			name:         "Fill Order Event",
			eventType:    string(orderBook.EventTypeFill),
			expectSave:   false,
			expectUpdate: true,
		},
		{
			name:         "Partial Fill Order Event",
			eventType:    string(orderBook.EventTypePartialFill),
			expectSave:   false,
			expectUpdate: true,
		},
		{
			name:         "Cancel Order Event",
			eventType:    string(orderBook.EventTypeCanceled),
			expectSave:   false,
			expectUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockOrderService)
			mockNotifier := new(MockEventNotifier)
			handler := NewOrderRequestHandler(mockService, mockNotifier)

			event := model.OrderEvent{
				EventType:   tt.eventType,
				OrderID:     "test-order-id",
				Price:       decimal.NewFromInt(100),
				Quantity:    decimal.NewFromInt(10),
				LeavesQty:   decimal.NewFromInt(5),
				ExecQty:     decimal.NewFromInt(5),
				Instrument:  "BTC/USDT",
				IsBid:       true,
				OrderStatus: tt.eventType,
				ExecType:    tt.eventType,
				Timestamp:   time.Now().UnixNano(),
			}

			eventJSON, _ := json.Marshal(event)

			if tt.expectSave {
				mockService.On("SaveOrderAsync", mock.MatchedBy(func(order model.Order) bool {
					return order.ID == event.OrderID &&
						order.Price.Equal(event.Price) &&
						order.OrderQty.Equal(event.Quantity)
				})).Return()
			}

			if tt.expectUpdate {
				mockService.On("UpdateOrderAsync",
					event.OrderID,
					event.OrderStatus,
					event.ExecType,
					event.LeavesQty,
					event.ExecQty,
				).Return()
			}

			err := handler.HandleEventMessages(eventJSON)
			assert.NoError(t, err)
			mockService.AssertExpectations(t)
		})
	}
}

func TestHandleMessage_WorkerCreation(t *testing.T) {
	mockService := new(MockOrderService)
	mockNotifier := new(MockEventNotifier)
	handler := NewOrderRequestHandler(mockService, mockNotifier)

	instrument := "BTC/USDT"
	orderReq := rmq.OrderRequest{
		RequestType: rmq.ReqTypeNew,
		Order: rmq.TraderOrder{
			ID:         "1",
			Price:      "100",
			Qty:        "10",
			Instrument: instrument,
			Side:       orderBook.Buy,
		},
	}

	body, _ := json.Marshal(orderReq)
	msg := amqp.Delivery{Body: body, Acknowledger: &mockAcknowledger{}}

	// First message should create a new worker
	handler.HandleMessage(context.Background(), msg)

	// Verify channel was created
	handler.mu.Lock()
	channel, exists := handler.orderChannels[instrument]
	handler.mu.Unlock()

	assert.True(t, exists)
	assert.NotNil(t, channel)

	// Let the worker process the message
	time.Sleep(100 * time.Millisecond)

	// Verify order book was created
	handler.mu.Lock()
	book, bookExists := handler.orderBooks[instrument]
	handler.mu.Unlock()

	assert.True(t, bookExists)
	assert.NotNil(t, book)
}

func TestHandleMessage_InvalidJSON(t *testing.T) {
	mockService := new(MockOrderService)
	mockNotifier := new(MockEventNotifier)
	handler := NewOrderRequestHandler(mockService, mockNotifier)

	msg := amqp.Delivery{
		Body:         []byte("invalid json"),
		Acknowledger: &mockAcknowledger{},
	}

	handler.HandleMessage(context.Background(), msg)
}

func TestWorkerProcessing(t *testing.T) {
	mockService := new(MockOrderService)
	mockNotifier := new(MockEventNotifier)
	handler := NewOrderRequestHandler(mockService, mockNotifier)

	instrument := "BTC/USDT"
	channel := make(chan rmq.OrderRequest, 100)
	book := orderBook.NewOrderBook(mockNotifier)

	handler.mu.Lock()
	handler.orderChannels[instrument] = channel
	handler.orderBooks[instrument] = book
	handler.mu.Unlock()

	// Start the worker
	go handler.startWorker(instrument, channel)

	// Send a new order request
	orderReq := rmq.OrderRequest{
		RequestType: rmq.ReqTypeNew,
		Order: rmq.TraderOrder{
			ID:         "1",
			Price:      "100",
			Qty:        "10",
			Instrument: instrument,
			Side:       orderBook.Buy,
		},
	}

	channel <- orderReq

	// Give the worker time to process
	time.Sleep(100 * time.Millisecond)

	// Verify the order book state
	handler.mu.Lock()
	book = handler.orderBooks[instrument]
	handler.mu.Unlock()

	assert.NotNil(t, book)
}
