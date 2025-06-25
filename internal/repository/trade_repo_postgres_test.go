package repository

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	sqlc "MatchingEngine/internal/db/sqlc"
	"MatchingEngine/internal/model"
)

// MockTradeQueries mocks the TradeQueries interface
type MockTradeQueries struct {
	mock.Mock
}

func (m *MockTradeQueries) CreateTrade(ctx context.Context, params sqlc.CreateTradeParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func (m *MockTradeQueries) CreateTradeSide(ctx context.Context, params sqlc.CreateTradeSideParams) error {
	args := m.Called(ctx, params)
	return args.Error(0)
}

func TestSaveTrade(t *testing.T) {
	mockQueries := new(MockTradeQueries)
	repo := NewPostgresTradeRepository(mockQueries)

	trade := model.TradeCaptureReport{
		MsgType:      "AE",
		ExecID:       "EXEC123",
		Symbol:       "BTC/USDT",
		LastQty:      decimal.NewFromInt(5),
		LastPx:       decimal.NewFromInt(100),
		TradeDate:    time.Now().Format("2006-01-02"),
		TransactTime: time.Now().UnixNano(),
		NoSides: []model.NoSides{
			{
				Side:    model.Buy,
				OrderID: "order-001",
			},
			{
				Side:    model.Sell,
				OrderID: "order-002",
			},
		},
	}

	// Match CreateTrade
	mockQueries.On("CreateTrade", mock.Anything, mock.MatchedBy(func(p sqlc.CreateTradeParams) bool {
		return p.ExecID == trade.ExecID && p.Symbol == trade.Symbol
	})).Return(nil)

	// Match CreateTradeSide for each side
	mockQueries.On("CreateTradeSide", mock.Anything, mock.MatchedBy(func(p sqlc.CreateTradeSideParams) bool {
		return p.OrderID == "order-001" || p.OrderID == "order-002"
	})).Return(nil).Twice()

	err := repo.SaveTrade(context.Background(), trade)
	assert.NoError(t, err)

	mockQueries.AssertExpectations(t)
}
