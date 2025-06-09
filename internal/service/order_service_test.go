package service

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"

	"MatchingEngine/internal/model"
	"MatchingEngine/internal/repository"
)

type MockAsyncDBWriter struct {
	mock.Mock
}

func (m *MockAsyncDBWriter) EnqueueTask(task repository.DBTask) {
	m.Called(task)
}

func TestSaveOrderAsync(t *testing.T) {
	mockWriter := new(MockAsyncDBWriter)
	orderService := NewOrderService(mockWriter)

	order := model.Order{
		ID:         "1",
		Instrument: "BTC/USDT",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}

	mockWriter.On("EnqueueTask", repository.SaveOrderTask{Order: order}).Return()

	orderService.SaveOrderAsync(order)

	mockWriter.AssertCalled(t, "EnqueueTask", repository.SaveOrderTask{Order: order})
}

func TestUpdateOrderAsync(t *testing.T) {
	mockWriter := new(MockAsyncDBWriter)
	orderService := NewOrderService(mockWriter)

	orderID := "1"
	leavesQty := decimal.NewFromInt(5)

	mockWriter.On("EnqueueTask", repository.UpdateOrderTask{OrderID: orderID, LeavesQty: leavesQty}).Return()

	orderService.UpdateOrderAsync(orderID, leavesQty)

	mockWriter.AssertCalled(t, "EnqueueTask", repository.UpdateOrderTask{OrderID: orderID, LeavesQty: leavesQty})
}

func TestCancelEventAsync(t *testing.T) {
	mockWriter := new(MockAsyncDBWriter)
	orderService := NewOrderService(mockWriter)

	event := model.Event{
		ID:         "event-1",
		OrderID:    "1",
		Instrument: "BTC/USDT",
		Type:       "canceled",
		OrderQty:   decimal.NewFromInt(10),
		LeavesQty:  decimal.Zero,
		Price:      decimal.NewFromInt(100),
	}

	mockWriter.On("EnqueueTask", repository.CancelEventTask{Event: event}).Return()

	orderService.CancelEventAsync(event)

	mockWriter.AssertCalled(t, "EnqueueTask", repository.CancelEventTask{Event: event})
}
