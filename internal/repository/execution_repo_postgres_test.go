package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	sqlc "MatchingEngine/internal/db/sqlc"
	"MatchingEngine/internal/model"
	"MatchingEngine/orderBook"
)

type MockQueries struct {
	mock.Mock
}

func (m *MockQueries) CreateExecution(ctx context.Context, params sqlc.CreateExecutionParams) (sqlc.Execution, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(sqlc.Execution), args.Error(1)
}


func TestSaveOrder(t *testing.T) {
	mockQueries := new(MockQueries)
	repo := NewPostgresOrderRepository(mockQueries)

	execReport := model.ExecutionReport{
		OrderID:          "1",
		Instrument:  "BTC/USDT",
		Price:       decimal.NewFromInt(100),
		OrderQty:    decimal.NewFromInt(10),
		LeavesQty:   decimal.NewFromInt(10),
		CumQty:      decimal.NewFromInt(0),
		OrderStatus: string(orderBook.OrderStatusNew),
		Timestamp:   time.Now().UnixNano(),
		IsBid:       true,
	}

	mockQueries.On("SaveExecution", mock.Anything, mock.Anything).Return(sqlc.Execution{
		OrderID:          execReport.OrderID,
		Instrument:  execReport.Instrument,
		Price:       pgtypeNumeric(execReport.Price),
		OrderQty:    pgtypeNumeric(execReport.OrderQty),
		LeavesQty:   pgtypeNumeric(execReport.LeavesQty),
		CumQty:      pgtypeNumeric(execReport.CumQty),
		OrderStatus: execReport.OrderStatus,
		Side:        "buy",
	}, nil)

	savedExecution, err := repo.SaveExecution(context.Background(), execReport)

	assert.NoError(t, err)
	assert.Equal(t, execReport.OrderID, savedExecution.OrderID)
	assert.Equal(t, execReport.Instrument, savedExecution.Instrument)
	assert.True(t, execReport.Price.Equal(savedExecution.Price))
	assert.True(t, execReport.OrderQty.Equal(savedExecution.OrderQty))
	assert.True(t, execReport.LeavesQty.Equal(savedExecution.LeavesQty))
	assert.Equal(t, execReport.OrderStatus, savedExecution.OrderStatus)
	mockQueries.AssertExpectations(t)
}

func pgtypeNumeric(d decimal.Decimal) pgtype.Numeric {
	var num pgtype.Numeric
	_ = num.Scan(d.String())
	return num
}
