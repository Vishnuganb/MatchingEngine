package service

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"

	"MatchingEngine/internal/repository"
	"MatchingEngine/orderBook"
)

type MockAsyncDBWriter struct {
	mock.Mock
}

func (m *MockAsyncDBWriter) EnqueueTask(task repository.DBTask) {
	m.Called(task)
}

func TestSaveOrderAsync(t *testing.T) {
	mockWriter := new(MockAsyncDBWriter)
	orderService := NewExecutionService(mockWriter)

	exec := orderBook.ExecutionReport{
		OrderID:    "1",
		Instrument: "BTC/USDT",
		Price:      decimal.NewFromInt(100),
		OrderQty:   decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}

	mockWriter.On("EnqueueTask", repository.SaveExecutionTask{Execution: exec}).Return()

	orderService.SaveExecutionAsync(exec)

	mockWriter.AssertCalled(t, "EnqueueTask", repository.SaveExecutionTask{Execution: exec})
}
