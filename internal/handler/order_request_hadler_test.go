package handler

import (
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

func (m *MockOrderService) SaveOrderAndEvent(ctx context.Context, order orderBook.Order, event orderBook.Event) (orderBook.Order, orderBook.Event, error) {
	args := m.Called(ctx, order, event)
	return args.Get(0).(orderBook.Order), args.Get(1).(orderBook.Event), args.Error(2)
}

func (m *MockOrderService) UpdateOrderAndEvent(ctx context.Context, orderID string, leavesQty decimal.Decimal, event orderBook.Event) error {
	args := m.Called(ctx, orderID, leavesQty, event)
	return args.Error(2)
}

func (m *MockOrderService) CancelEvent(ctx context.Context, event orderBook.Event) error {
	args := m.Called(ctx, event)
	return args.Error(1)
}

func TestHandleNewOrder(t *testing.T) {
	mockService := new(MockOrderService)
	book := orderBook.NewOrderBook()
	_ = NewOrderRequestHandler(book, mockService)

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
	book := orderBook.NewOrderBook()
	_ = NewOrderRequestHandler(book, mockService)

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

	book.NewOrder(order)

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
	orderBook := orderBook.NewOrderBook()
	handler := NewOrderRequestHandler(orderBook, mockService)

	msg := amqp.Delivery{Body: []byte("invalid message")}

	handler.HandleMessage(context.Background(), msg)
}

func TestHandleUnknownRequestType(t *testing.T) {
	mockService := new(MockOrderService)
	orderBook := orderBook.NewOrderBook()
	handler := NewOrderRequestHandler(orderBook, mockService)

	req := rmq.OrderRequest{
		RequestType: 999, // Unknown request type
	}

	body, _ := json.Marshal(req)
	msg := amqp.Delivery{Body: body}

	handler.HandleMessage(context.Background(), msg)
}
