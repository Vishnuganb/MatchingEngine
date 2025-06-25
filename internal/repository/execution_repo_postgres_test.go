package repository

import (
	"MatchingEngine/internal/model"
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	sqlc "MatchingEngine/internal/db/sqlc"
)

type MockQueries struct {
	mock.Mock
}

func (m *MockQueries) CreateExecution(ctx context.Context, params sqlc.CreateExecutionParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func TestSaveExecution(t *testing.T) {
	mockQueries := new(MockQueries)
	repo := NewPostgresExecutionRepository(mockQueries)

	execReport := model.ExecutionReport{
		MsgType:      "8",
		ExecID:       "exec-123",
		OrderID:      "order-1",
		ClOrdID:      "CL001",
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

	mockQueries.
		On("CreateExecution", mock.Anything, mock.Anything).
		Return(nil)

	err := repo.SaveExecution(context.Background(), execReport)
	assert.NoError(t, err)

	mockQueries.AssertExpectations(t)
}
