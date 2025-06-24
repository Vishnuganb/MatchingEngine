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

func TestSaveExecutionAsync(t *testing.T) {
	mockWriter := new(MockAsyncDBWriter)
	executionService := NewExecutionService(mockWriter)

	execution := model.ExecutionReport{
		ExecID:       "execution-123456",
		OrderID:      "order-123456",
		ClOrdID:      "clord-123",
		ExecType:     model.ExecTypeFill,
		OrdStatus:    model.OrderStatusPartialFill,
		Symbol:       "BTC/USDT",
		Side:         model.Buy,
		OrderQty:     decimal.NewFromInt(10),
		LastShares:   decimal.NewFromInt(5),
		LastPx:       decimal.NewFromInt(100),
		LeavesQty:    decimal.NewFromInt(5),
		CumQty:       decimal.NewFromInt(5),
		AvgPx:        decimal.NewFromInt(100),
		TransactTime: time.Now().UnixNano(),
		Text:         "Trade executed",
	}

	mockWriter.On("EnqueueTask", repository.SaveExecutionTask{Execution: execution}).Return()

	executionService.SaveExecutionAsync(execution)

	mockWriter.AssertCalled(t, "EnqueueTask", repository.SaveExecutionTask{Execution: execution})
}
