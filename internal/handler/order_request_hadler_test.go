package handler

import (
	"MatchingEngine/internal/model"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/mock"

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

// MockAcknowledger simulates the Acknowledger interface for testing
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

func TestHandleNewOrder(t *testing.T) {
	mockService := new(MockOrderService)
	handler := NewOrderRequestHandler(mockService)

	orderReq := rmq.OrderRequest{
		RequestType: rmq.ReqTypeNew,
		Order: rmq.TraderOrder{
			ID:         "1",
			Price:      "100",
			Qty:        "10",
			Instrument: "BTC/USDT",
			Side:       orderBook.Buy,
		},
	}

	order := model.Order{
		ID:         "1",
		Price:      decimal.RequireFromString("100"),
		OrderQty:   decimal.RequireFromString("10"),
		Instrument: "BTC/USDT",
		IsBid:      true,
	}

	mockService.On("SaveOrderAsync", mock.MatchedBy(func(o model.Order) bool {
		return o.ID == order.ID
	})).Return()

	body, _ := json.Marshal(orderReq)
	msg := amqp.Delivery{Body: body, Acknowledger: &mockAcknowledger{}}

	handler.HandleMessage(context.Background(), msg)

	// Wait for goroutine to process
	time.Sleep(100 * time.Millisecond)

	mockService.AssertExpectations(t)

}

func TestHandleCancelOrder(t *testing.T) {
	mockService := new(MockOrderService)
	handler := NewOrderRequestHandler(mockService)

	// Set up the mock expectation first
	mockService.On("UpdateOrderAsync",
		"1",                  // orderID
		"",           // orderStatus
		"canceled",           // execType
		decimal.Zero,         // leavesQty
		decimal.Zero,         // execQty
	).Return()

	// Create and initialize the order book
	book := orderBook.NewOrderBook()
	order := model.Order{
		ID:          "1",
		Instrument:  "BTC/USDT",
		Price:       decimal.NewFromInt(100),
		OrderQty:    decimal.NewFromInt(10),
		LeavesQty:   decimal.NewFromInt(10),
		Timestamp:   time.Now().UnixNano(),
		IsBid:       true,
		OrderStatus: "new",
		ExecType:    "new",
		ExecQty:     decimal.Zero,
	}

	// Set the order book and service in the handler
	handler.mu.Lock()
	handler.orderBooks["BTC/USDT"] = book
	handler.OrderService = mockService
	handler.mu.Unlock()

	// Add the order to the book
	book.OnNewOrder(order)

	// Create and send the cancel request
	cancelReq := rmq.OrderRequest{
		RequestType: rmq.ReqTypeCancel,
		Order: rmq.TraderOrder{
			ID:         "1",
			Instrument: "BTC/USDT",
		},
	}

	body, _ := json.Marshal(cancelReq)
	msg := amqp.Delivery{Body: body, Acknowledger: &mockAcknowledger{}}

	handler.HandleMessage(context.Background(), msg)

	// Wait for goroutine to process
	time.Sleep(100 * time.Millisecond)

	mockService.AssertExpectations(t)
}

func TestHandleInvalidMessage(t *testing.T) {
	mockService := new(MockOrderService)
	handler := NewOrderRequestHandler(mockService)

	msg := amqp.Delivery{Body: []byte("invalid message")}

	handler.HandleMessage(context.Background(), msg)
}

func TestHandleUnknownRequestType(t *testing.T) {
	mockService := new(MockOrderService)
	handler := NewOrderRequestHandler(mockService)

	req := rmq.OrderRequest{
		RequestType: 999,
		Order: rmq.TraderOrder{
			ID:         "unknown",
			Instrument: "BTC/USDT",
		},
	}

	body, _ := json.Marshal(req)
	msg := amqp.Delivery{Body: body, Acknowledger: &mockAcknowledger{}}

	handler.HandleMessage(context.Background(), msg)
}
