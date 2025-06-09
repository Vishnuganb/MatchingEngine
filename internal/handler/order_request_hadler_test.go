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

func (m *MockOrderService) SaveEventAsync(event model.Event) {
	m.Called(event)
}

func (m *MockOrderService) CancelEventAsync(event model.Event) {
	m.Called(event)
}

func (m *MockOrderService) UpdateOrderAsync(orderID string, leavesQty decimal.Decimal) {
	m.Called(orderID, leavesQty)
}

func TestHandleNewOrder(t *testing.T) {
	mockService := new(MockOrderService)
	_ = NewOrderRequestHandler(mockService)

	order := orderBook.Order{
		ID:         "1",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		Instrument: "BTC/USDT",
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}

	event := orderBook.Event{
		ID:         "event-1",
		OrderID:    "1",
		Instrument: "BTC/USDT",
		Type:       orderBook.EventTypeNew,
		OrderQty:   decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
		Price:      decimal.NewFromInt(100),
	}

	mockService.On("SaveOrderAndEvent", mock.Anything, mock.MatchedBy(func(o orderBook.Order) bool {
		return o.ID == order.ID && o.Instrument == order.Instrument
	}), mock.MatchedBy(func(e orderBook.Event) bool {
		return e.OrderID == event.OrderID && e.Type == event.Type
	})).Return(order, event, nil)
	mockService.On("UpdateOrderAndEvent", mock.Anything, order.ID, order.LeavesQty, mock.Anything).Return(order, event, nil)

}

func TestHandleCancelOrder(t *testing.T) {
	mockService := new(MockOrderService)
	_ = NewOrderRequestHandler(mockService)

	// Initialize the OrderBook
	book := orderBook.NewOrderBook()

	// Step 1: Create an order in memory
	order := orderBook.Order{
		ID:         "1",
		Instrument: "BTC/USDT",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}

	book.AddBuyOrder(order)

	// Step 2: Cancel the order
	cancelRequest := orderBook.OrderRequest{
		ID:   "1",
		Side: orderBook.Buy,
	}

	_ = book.CancelOrder(cancelRequest.ID)

	// Create a valid event object
	expectedEvent := orderBook.Event{
		ID:         "event-1",
		OrderID:    "1",
		Instrument: "BTC/USDT",
		Type:       orderBook.EventTypeCanceled,
		Price:      decimal.NewFromInt(100),
		OrderQty:   decimal.NewFromInt(10),
		LeavesQty:  decimal.Zero,
		ExecQty:    decimal.Zero,
	}

	mockService.On("CancelEvent", mock.Anything, mock.MatchedBy(func(e orderBook.Event) bool {
		return e.OrderID == "1" && e.Type == orderBook.EventTypeCanceled
	})).Return(expectedEvent, nil)
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
		RequestType: 999, // Unknown request type
	}

	body, _ := json.Marshal(req)
	msg := amqp.Delivery{Body: body}

	handler.HandleMessage(context.Background(), msg)
}
