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

func (m *MockQueries) CreateActiveOrder(ctx context.Context, params sqlc.CreateActiveOrderParams) (sqlc.ActiveOrder, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(sqlc.ActiveOrder), args.Error(1)
}

func (m *MockQueries) UpdateActiveOrder(ctx context.Context, params sqlc.UpdateActiveOrderParams) (sqlc.ActiveOrder, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(sqlc.ActiveOrder), args.Error(1)
}

func TestSaveOrder(t *testing.T) {
	mockQueries := new(MockQueries)
	repo := NewPostgresOrderRepository(mockQueries)

	order := model.Order{
		ID:          "1",
		Instrument:  "BTC/USDT",
		Price:       decimal.NewFromInt(100),
		OrderQty:    decimal.NewFromInt(10),
		LeavesQty:   decimal.NewFromInt(10),
		CumQty:      decimal.NewFromInt(0),
		OrderStatus: string(orderBook.OrderStatusNew),
		Timestamp:   time.Now().UnixNano(),
		IsBid:       true,
	}

	mockQueries.On("CreateActiveOrder", mock.Anything, mock.Anything).Return(sqlc.ActiveOrder{
		ID:          order.ID,
		Instrument:  order.Instrument,
		Price:       pgtypeNumeric(order.Price),
		OrderQty:    pgtypeNumeric(order.OrderQty),
		LeavesQty:   pgtypeNumeric(order.LeavesQty),
		CumQty:      pgtypeNumeric(order.CumQty),
		OrderStatus: order.OrderStatus,
		Side:        "buy",
	}, nil)

	savedOrder, err := repo.SaveOrder(context.Background(), order)

	assert.NoError(t, err)
	assert.Equal(t, order.ID, savedOrder.ID)
	assert.Equal(t, order.Instrument, savedOrder.Instrument)
	assert.True(t, order.Price.Equal(savedOrder.Price))
	assert.True(t, order.OrderQty.Equal(savedOrder.OrderQty))
	assert.True(t, order.LeavesQty.Equal(savedOrder.LeavesQty))
	assert.Equal(t, order.OrderStatus, savedOrder.OrderStatus)
	mockQueries.AssertExpectations(t)
}

func TestUpdateOrder(t *testing.T) {
	mockQueries := new(MockQueries)
	repo := NewPostgresOrderRepository(mockQueries)

	orderID := "1"
	orderStatus := string(orderBook.OrderStatusFill)
	leavesQty := decimal.NewFromInt(5)
	cumQty := decimal.NewFromInt(10)
	price := decimal.NewFromInt(100)

	mockQueries.On("UpdateActiveOrder", mock.Anything, mock.Anything).Return(sqlc.ActiveOrder{
		ID:          orderID,
		LeavesQty:   pgtypeNumeric(leavesQty),
		CumQty:      pgtypeNumeric(cumQty),
		Price:       pgtypeNumeric(price),
		OrderQty:    pgtypeNumeric(decimal.NewFromInt(15)),
		OrderStatus: orderStatus,
	}, nil)

	updatedOrder, err := repo.UpdateOrder(context.Background(), orderID, orderStatus, leavesQty, cumQty, price)

	assert.NoError(t, err)
	assert.Equal(t, orderID, updatedOrder.ID)
	assert.True(t, leavesQty.Equal(updatedOrder.LeavesQty))
	assert.True(t, cumQty.Equal(updatedOrder.CumQty))
	assert.True(t, price.Equal(updatedOrder.Price))
	assert.Equal(t, orderStatus, updatedOrder.OrderStatus)
	mockQueries.AssertExpectations(t)
}

func pgtypeNumeric(d decimal.Decimal) pgtype.Numeric {
	var num pgtype.Numeric
	_ = num.Scan(d.String())
	return num
}
