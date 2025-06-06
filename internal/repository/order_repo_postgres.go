package repository

import (
	"MatchingEngine/internal/model"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	sqlc "MatchingEngine/internal/db/sqlc"
	"MatchingEngine/orderBook"
)

type OrderQueries interface {
	CreateActiveOrder(ctx context.Context, params sqlc.CreateActiveOrderParams) (sqlc.ActiveOrder, error)
	UpdateActiveOrder(ctx context.Context, params sqlc.UpdateActiveOrderParams) (sqlc.ActiveOrder, error)
	CreateEvent(ctx context.Context, params sqlc.CreateEventParams) (sqlc.Event, error)
	UpdateEvent(ctx context.Context, params sqlc.UpdateEventParams) (sqlc.Event, error)
}

type PostgresOrderRepository struct {
	queries OrderQueries
}

func NewPostgresOrderRepository(queries OrderQueries) *PostgresOrderRepository {
	return &PostgresOrderRepository{queries: queries}
}

func decimalToPgNumeric(d decimal.Decimal) (pgtype.Numeric, error) {
	var num pgtype.Numeric
	err := num.Scan(d.String())
	return num, err
}

func (r *PostgresOrderRepository) SaveOrder(ctx context.Context, order model.Order) (model.Order, error) {
	qty, err := decimalToPgNumeric(order.Qty)
	if err != nil {
		return model.Order{}, err
	}
	leavesQty, err := decimalToPgNumeric(order.LeavesQty)
	if err != nil {
		return model.Order{}, err
	}
	price, err := decimalToPgNumeric(order.Price)
	if err != nil {
		return model.Order{}, err
	}

	activeOrder, err := r.queries.CreateActiveOrder(ctx, sqlc.CreateActiveOrderParams{
		ID:         order.ID,
		Qty:        qty,
		LeavesQty:  leavesQty,
		Price:      price,
		Instrument: order.Instrument,
	})
	if err != nil {
		return model.Order{}, err
	}

	mappedOrder, err := MapActiveOrderToOrder(activeOrder)
	if err != nil {
		return model.Order{}, err
	}

	return mappedOrder, nil
}

func (r *PostgresOrderRepository) UpdateOrder(ctx context.Context, orderID string, leavesQty decimal.Decimal) (model.Order, error) {
	leavesQtyNumeric, err := decimalToPgNumeric(leavesQty)
	if err != nil {
		return model.Order{}, err
	}

	activeOrder, err := r.queries.UpdateActiveOrder(ctx, sqlc.UpdateActiveOrderParams{
		ID:        orderID,
		LeavesQty: leavesQtyNumeric,
	})
	if err != nil {
		return model.Order{}, err
	}

	mappedOrder, err := MapActiveOrderToOrder(activeOrder)
	if err != nil {
		return model.Order{}, err
	}

	return mappedOrder, nil
}

func (r *PostgresOrderRepository) SaveEvent(ctx context.Context, event model.Event) (model.Event, error) {
	orderQty, err := decimalToPgNumeric(event.OrderQty)
	if err != nil {
		return model.Event{}, err
	}
	leavesQty, err := decimalToPgNumeric(event.LeavesQty)
	if err != nil {
		return model.Event{}, err
	}
	execQty, err := decimalToPgNumeric(event.ExecQty)
	if err != nil {
		return model.Event{}, err
	}
	price, err := decimalToPgNumeric(event.Price)
	if err != nil {
		return model.Event{}, err
	}

	dbEvent, err := r.queries.CreateEvent(ctx, sqlc.CreateEventParams{
		OrderID:    event.OrderID,
		Type:       event.Type,
		Side:       event.Side,
		OrderQty:   orderQty,
		LeavesQty:  leavesQty,
		ExecQty:    execQty,
		Price:      price,
		Instrument: event.Instrument,
	})
	if err != nil {
		return model.Event{}, err
	}

	return MapDBEventToOrderEvent(dbEvent), nil
}

func (r *PostgresOrderRepository) UpdateEvent(ctx context.Context, event model.Event) (model.Event, error) {
	orderQty, err := decimalToPgNumeric(event.OrderQty)
	if err != nil {
		return model.Event{}, err
	}
	leavesQty, err := decimalToPgNumeric(event.LeavesQty)
	if err != nil {
		return model.Event{}, err
	}
	execQty, err := decimalToPgNumeric(event.ExecQty)
	if err != nil {
		return model.Event{}, err
	}
	price, err := decimalToPgNumeric(event.Price)
	if err != nil {
		return model.Event{}, err
	}

	eventID, err := uuid.Parse(event.ID)
	if err != nil {
		return model.Event{}, fmt.Errorf("invalid UUID for event ID: %w", err)
	}

	dbEvent, err := r.queries.UpdateEvent(ctx, sqlc.UpdateEventParams{
		ID:        eventID,
		OrderID:   pgtype.Text{String: event.OrderID, Valid: true},
		Type:      pgtype.Text{String: event.Type, Valid: true},
		Side:      pgtype.Text{String: event.Side, Valid: true},
		OrderQty:  orderQty,
		LeavesQty: leavesQty,
		ExecQty:   execQty,
		Price:     price,
	})
	if err != nil {
		return model.Event{}, err
	}

	return MapDBEventToOrderEvent(dbEvent), nil
}

func MapDBEventToOrderEvent(dbEvent sqlc.Event) model.Event {
	orderQty, err := pgNumericToDecimal(dbEvent.OrderQty)
	if err != nil {
		orderQty = decimal.Zero // Default to zero if conversion fails
	}
	leavesQty, err := pgNumericToDecimal(dbEvent.LeavesQty)
	if err != nil {
		leavesQty = decimal.Zero
	}
	execQty, err := pgNumericToDecimal(dbEvent.ExecQty)
	if err != nil {
		execQty = decimal.Zero
	}
	price, err := pgNumericToDecimal(dbEvent.Price)
	if err != nil {
		price = decimal.Zero
	}

	return model.Event{
		ID:         dbEvent.ID.String(),
		OrderID:    dbEvent.OrderID,
		Timestamp:  dbEvent.Timestamp.UnixNano(),
		Type:       dbEvent.Type,
		Side:       dbEvent.Side,
		Instrument: dbEvent.Instrument,
		OrderQty:   orderQty,
		LeavesQty:  leavesQty,
		ExecQty:    execQty,
		Price:      price,
	}
}

func MapActiveOrderToOrder(activeOrder sqlc.ActiveOrder) (model.Order, error) {
	price, err := pgNumericToDecimal(activeOrder.Price)
	if err != nil {
		return model.Order{}, fmt.Errorf("converting price: %w", err)
	}
	qty, err := pgNumericToDecimal(activeOrder.Qty)
	if err != nil {
		return model.Order{}, fmt.Errorf("converting qty: %w", err)
	}
	leavesQty, err := pgNumericToDecimal(activeOrder.LeavesQty)
	if err != nil {
		return model.Order{}, fmt.Errorf("converting leavesQty: %w", err)
	}

	return model.Order{
		ID:         activeOrder.ID,
		Price:      price,
		Qty:        qty,
		Instrument: activeOrder.Instrument,
		LeavesQty:  leavesQty,
		IsBid:      activeOrder.Side == string(orderBook.Buy),
	}, nil
}

func pgNumericToDecimal(num pgtype.Numeric) (decimal.Decimal, error) {
	val, err := num.Value()
	if err != nil {
		return decimal.Decimal{}, err
	}
	str, ok := val.(string)
	if !ok {
		return decimal.Decimal{}, fmt.Errorf("unexpected type for pgtype.Numeric: %T", val)
	}
	dec, err := decimal.NewFromString(str)
	if err != nil {
		return decimal.Decimal{}, err
	}
	return dec, nil
}
