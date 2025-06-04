package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"MatchingEngine/orderBook"
)

// MockOrderRepository is a mock implementation of the OrderRepository interface
type MockOrderRepository struct {
	mock.Mock
}

type MockKafkaProducer struct {
	mock.Mock
}

func (m *MockOrderRepository) SaveOrder(ctx context.Context, order orderBook.Order) (orderBook.Order, error) {
	args := m.Called(ctx, order)
	return args.Get(0).(orderBook.Order), args.Error(1)
}

func (m *MockOrderRepository) SaveEvent(ctx context.Context, event orderBook.Event) (orderBook.Event, error) {
	args := m.Called(ctx, event)
	return args.Get(0).(orderBook.Event), args.Error(1)
}

func (m *MockOrderRepository) UpdateOrder(ctx context.Context, orderID string, leavesQty decimal.Decimal) (orderBook.Order, error) {
	args := m.Called(ctx, orderID, leavesQty)
	return args.Get(0).(orderBook.Order), args.Error(1)
}

func (m *MockOrderRepository) UpdateEvent(ctx context.Context, event orderBook.Event) (orderBook.Event, error) {
	args := m.Called(ctx, event)
	return args.Get(0).(orderBook.Event), args.Error(1)
}

func (m *MockKafkaProducer) NotifyEventAndOrder(key string, value json.RawMessage) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func TestSaveOrderAndEvent(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	mockProducer := new(MockKafkaProducer)
	orderService := NewOrderService(mockRepo, mockProducer)

	order := orderBook.Order{
		ID:         "1",
		Instrument: "BTC/USDT",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
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

	mockRepo.On("SaveOrder", mock.Anything, order).Return(order, nil)
	mockRepo.On("SaveEvent", mock.Anything, event).Return(event, nil)

	savedOrder, savedEvent, err := orderService.SaveOrderAndEvent(context.Background(), order, event)

	assert.NoError(t, err)
	assert.Equal(t, order, savedOrder)
	assert.Equal(t, event, savedEvent)
	mockRepo.AssertExpectations(t)
}

func TestUpdateOrderAndEvent(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	mockProducer := new(MockKafkaProducer)
	orderService := NewOrderService(mockRepo, mockProducer)

	orderID := "1"
	leavesQty := decimal.NewFromInt(5)
	order := orderBook.Order{
		ID:         orderID,
		Instrument: "BTC/USDT",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		LeavesQty:  leavesQty,
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}

	event := orderBook.Event{
		ID:         "event-1",
		OrderID:    orderID,
		Instrument: "BTC/USDT",
		Type:       orderBook.EventTypeNew,
		OrderQty:   decimal.NewFromInt(10),
		LeavesQty:  leavesQty,
		Price:      decimal.NewFromInt(100),
	}

	updatedEvent := event
	updatedEvent.Type = orderBook.EventTypePartialFill

	mockRepo.On("UpdateOrder", mock.Anything, orderID, leavesQty).Return(order, nil)
	mockRepo.On("UpdateEvent", mock.Anything, updatedEvent).Return(updatedEvent, nil)

	err := orderService.UpdateOrderAndEvent(context.Background(), orderID, leavesQty, updatedEvent)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCancelEvent(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	mockProducer := new(MockKafkaProducer)
	orderService := NewOrderService(mockRepo, mockProducer)

	event := orderBook.Event{
		ID:         "event-1",
		OrderID:    "1",
		Instrument: "BTC/USDT",
		Type:       orderBook.EventTypeCanceled,
		OrderQty:   decimal.NewFromInt(10),
		LeavesQty:  decimal.Zero,
		Price:      decimal.NewFromInt(100),
	}

	mockRepo.On("SaveEvent", mock.Anything, event).Return(event, nil)

	err := orderService.CancelEvent(context.Background(), event)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSaveOrderAndEvent_Failure(t *testing.T) {
	mockRepo := new(MockOrderRepository)
	mockProducer := new(MockKafkaProducer)
	orderService := NewOrderService(mockRepo, mockProducer)

	order := orderBook.Order{
		ID:         "1",
		Instrument: "BTC/USDT",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
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

	mockRepo.On("SaveOrder", mock.Anything, order).Return(orderBook.Order{}, errors.New("database error"))

	_, _, err := orderService.SaveOrderAndEvent(context.Background(), order, event)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
