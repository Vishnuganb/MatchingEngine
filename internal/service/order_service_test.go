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
		OrderQty:   decimal.NewFromInt(10),
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
	execQty := decimal.NewFromInt(10)
	orderStatus := "partially_fill"
	execType := "fill"

	mockWriter.On("EnqueueTask", repository.UpdateOrderTask{OrderID: orderID, LeavesQty: leavesQty, OrderStatus: orderStatus, ExecQty: execQty, ExecType: execType}).Return()

	orderService.UpdateOrderAsync(orderID, orderStatus, execType, leavesQty, execQty)

	mockWriter.AssertCalled(t, "EnqueueTask", repository.UpdateOrderTask{OrderID: orderID, LeavesQty: leavesQty, OrderStatus: orderStatus, ExecQty: execQty, ExecType: execType})
}
