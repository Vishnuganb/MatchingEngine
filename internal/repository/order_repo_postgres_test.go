package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	sqlc "MatchingEngine/internal/db/sqlc"
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

func (m *MockQueries) CreateEvent(ctx context.Context, params sqlc.CreateEventParams) (sqlc.Event, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(sqlc.Event), args.Error(1)
}

func (m *MockQueries) UpdateEvent(ctx context.Context, params sqlc.UpdateEventParams) (sqlc.Event, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(sqlc.Event), args.Error(1)
}

func TestSaveOrder(t *testing.T) {
	mockQueries := new(MockQueries)
	repo := NewPostgresOrderRepository(mockQueries)

	order := orderBook.Order{
		ID:         "1",
		Instrument: "BTC/USDT",
		Price:      decimal.NewFromInt(100),
		Qty:        decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
		Timestamp:  time.Now().UnixNano(),
		IsBid:      true,
	}

	mockQueries.On("CreateActiveOrder", mock.Anything, mock.Anything).Return(sqlc.ActiveOrder{
		ID:         order.ID,
		Instrument: order.Instrument,
		Price:      pgtypeNumeric(order.Price),
		Qty:        pgtypeNumeric(order.Qty),
		LeavesQty:  pgtypeNumeric(order.LeavesQty),
		Side:       "buy",
	}, nil)

	savedOrder, err := repo.SaveOrder(context.Background(), order)

	assert.NoError(t, err)
	assert.Equal(t, order.ID, savedOrder.ID)
	mockQueries.AssertExpectations(t)
}

func TestUpdateOrder(t *testing.T) {
	mockQueries := new(MockQueries)
	repo := NewPostgresOrderRepository(mockQueries)

	orderID := "1"
	leavesQty := decimal.NewFromInt(5)

	mockQueries.On("UpdateActiveOrder", mock.Anything, mock.Anything).Return(sqlc.ActiveOrder{
		ID:        orderID,
		Qty:       pgtypeNumeric(decimal.NewFromInt(10)),
		Price:     pgtypeNumeric(decimal.NewFromInt(100)),
		LeavesQty: pgtypeNumeric(leavesQty),
	}, nil)

	updatedOrder, err := repo.UpdateOrder(context.Background(), orderID, leavesQty)

	assert.NoError(t, err)
	assert.Equal(t, orderID, updatedOrder.ID)
	assert.Equal(t, leavesQty, updatedOrder.LeavesQty)
	mockQueries.AssertExpectations(t)
}

func TestSaveEvent(t *testing.T) {
	mockQueries := new(MockQueries)
	repo := NewPostgresOrderRepository(mockQueries)

	event := orderBook.Event{
		ID:         uuid.New().String(),
		OrderID:    "1",
		Instrument: "BTC/USDT",
		Type:       orderBook.EventTypeNew,
		OrderQty:   decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(10),
		Price:      decimal.NewFromInt(100),
	}

	mockQueries.On("CreateEvent", mock.Anything, mock.Anything).Return(sqlc.Event{
		ID:         uuid.MustParse(event.ID),
		OrderID:    event.OrderID,
		Instrument: event.Instrument,
		Type:       string(event.Type),
		OrderQty:   pgtypeNumeric(event.OrderQty),
		LeavesQty:  pgtypeNumeric(event.LeavesQty),
		Price:      pgtypeNumeric(event.Price),
	}, nil)

	savedEvent, err := repo.SaveEvent(context.Background(), event)

	assert.NoError(t, err)
	assert.Equal(t, event.ID, savedEvent.ID)
	mockQueries.AssertExpectations(t)
}

func TestUpdateEvent(t *testing.T) {
	mockQueries := new(MockQueries)
	repo := NewPostgresOrderRepository(mockQueries)

	event := orderBook.Event{
		ID:         uuid.New().String(),
		OrderID:    "1",
		Instrument: "BTC/USDT",
		Type:       orderBook.EventTypePartialFill,
		OrderQty:   decimal.NewFromInt(10),
		LeavesQty:  decimal.NewFromInt(5),
		Price:      decimal.NewFromInt(100),
	}

	mockQueries.On("UpdateEvent", mock.Anything, mock.Anything).Return(sqlc.Event{
		ID:         uuid.MustParse(event.ID),
		OrderID:    event.OrderID,
		Instrument: event.Instrument,
		Type:       string(event.Type),
		OrderQty:   pgtypeNumeric(event.OrderQty),
		LeavesQty:  pgtypeNumeric(event.LeavesQty),
		Price:      pgtypeNumeric(event.Price),
	}, nil)

	updatedEvent, err := repo.UpdateEvent(context.Background(), event)

	assert.NoError(t, err)
	assert.Equal(t, event.ID, updatedEvent.ID)
	assert.Equal(t, event.Type, updatedEvent.Type)
	mockQueries.AssertExpectations(t)
}

func pgtypeNumeric(d decimal.Decimal) pgtype.Numeric {
	var num pgtype.Numeric
	_ = num.Scan(d.String())
	return num
}
