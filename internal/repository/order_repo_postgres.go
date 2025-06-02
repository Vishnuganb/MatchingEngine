package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	sqlc "MatchingEngine/internal/db/sqlc"
	"MatchingEngine/orderBook"
)

type PostgresOrderRepository struct {
	queries *sqlc.Queries
}

func NewPostgresOrderRepository(queries *sqlc.Queries) *PostgresOrderRepository {
	return &PostgresOrderRepository{queries: queries}
}

func decimalToPgNumeric(d decimal.Decimal) (pgtype.Numeric, error) {
	var num pgtype.Numeric
	err := num.Scan(d.String())
	return num, err
}

func (r *PostgresOrderRepository) SaveOrder(ctx context.Context, order orderBook.Order) (orderBook.Order, error) {
	qty, err := decimalToPgNumeric(order.Qty)
	if err != nil {
		return orderBook.Order{}, err
	}
	leavesQty, err := decimalToPgNumeric(order.LeavesQty)
	if err != nil {
		return orderBook.Order{}, err
	}
	price, err := decimalToPgNumeric(order.Price)
	if err != nil {
		return orderBook.Order{}, err
	}

	activeOrder, err := r.queries.CreateActiveOrder(ctx, sqlc.CreateActiveOrderParams{
		ID:         order.ID,
		Side:       string(order.Side()),
		Qty:        qty,
		LeavesQty:  leavesQty,
		Price:      price,
		Instrument: order.Instrument,
	})
	if err != nil {
		return orderBook.Order{}, err
	}

	mappedOrder, err := MapActiveOrderToOrder(activeOrder)
	if err != nil {
		return orderBook.Order{}, err
	}

	return mappedOrder, nil
}

func (r *PostgresOrderRepository) UpdateOrder(ctx context.Context, orderID string, leavesQty decimal.Decimal) (orderBook.Order, error) {
	leavesQtyNumeric, err := decimalToPgNumeric(leavesQty)
	if err != nil {
		return orderBook.Order{}, err
	}

	activeOrder, err := r.queries.UpdateActiveOrder(ctx, sqlc.UpdateActiveOrderParams{
		ID:        orderID,
		LeavesQty: leavesQtyNumeric,
	})
	if err != nil {
		return orderBook.Order{}, err
	}

	mappedOrder, err := MapActiveOrderToOrder(activeOrder)
	if err != nil {
		return orderBook.Order{}, err
	}

	return mappedOrder, nil
}

func (r *PostgresOrderRepository) SaveEvent(ctx context.Context, event orderBook.Event) (orderBook.Event, error) {
	orderQty, err := decimalToPgNumeric(event.OrderQty)
	if err != nil {
		return orderBook.Event{}, err
	}
	leavesQty, err := decimalToPgNumeric(event.LeavesQty)
	if err != nil {
		return orderBook.Event{}, err
	}
	execQty, err := decimalToPgNumeric(event.ExecQty)
	if err != nil {
		return orderBook.Event{}, err
	}
	price, err := decimalToPgNumeric(event.Price)
	if err != nil {
		return orderBook.Event{}, err
	}

	dbEvent, err := r.queries.CreateEvent(ctx, sqlc.CreateEventParams{
		OrderID:    event.OrderID,
		Type:       string(event.Type),
		Side:       string(event.Side),
		OrderQty:   orderQty,
		LeavesQty:  leavesQty,
		ExecQty:    execQty,
		Price:      price,
		Instrument: event.Instrument,
	})
	if err != nil {
		return orderBook.Event{}, err
	}

	return MapDBEventToOrderEvent(dbEvent), nil
}

func (r *PostgresOrderRepository) UpdateEvent(ctx context.Context, event orderBook.Event) (orderBook.Event, error) {
	orderQty, err := decimalToPgNumeric(event.OrderQty)
	if err != nil {
		return orderBook.Event{}, err
	}
	leavesQty, err := decimalToPgNumeric(event.LeavesQty)
	if err != nil {
		return orderBook.Event{}, err
	}
	execQty, err := decimalToPgNumeric(event.ExecQty)
	if err != nil {
		return orderBook.Event{}, err
	}
	price, err := decimalToPgNumeric(event.Price)
	if err != nil {
		return orderBook.Event{}, err
	}

	eventID, err := uuid.Parse(event.ID)
	if err != nil {
		return orderBook.Event{}, fmt.Errorf("invalid UUID for event ID: %w", err)
	}

	dbEvent, err := r.queries.UpdateEvent(ctx, sqlc.UpdateEventParams{
		ID:        eventID,
		OrderID:   pgtype.Text{String: string(event.OrderID), Valid: true},
		Type:      pgtype.Text{String: string(event.Type), Valid: true},
		Side:      pgtype.Text{String: string(event.Side), Valid: true},
		OrderQty:  orderQty,
		LeavesQty: leavesQty,
		ExecQty:   execQty,
		Price:     price,
	})
	if err != nil {
		return orderBook.Event{}, err
	}

	return MapDBEventToOrderEvent(dbEvent), nil
}

func MapDBEventToOrderEvent(dbEvent sqlc.Event) orderBook.Event {
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

	return orderBook.Event{
		ID:        dbEvent.ID.String(),
		OrderID:   dbEvent.OrderID,
		Timestamp: dbEvent.Timestamp.UnixNano(),
		Type:      orderBook.EventType(dbEvent.Type),
		Side:      orderBook.Side(dbEvent.Side),
		Instrument: dbEvent.Instrument,
		OrderQty:  orderQty,
		LeavesQty: leavesQty,
		ExecQty:   execQty,
		Price:     price,
	}
}

func MapActiveOrderToOrder(activeOrder sqlc.ActiveOrder) (orderBook.Order, error) {
	price, err := pgNumericToDecimal(activeOrder.Price)
	if err != nil {
		return orderBook.Order{}, fmt.Errorf("converting price: %w", err)
	}
	qty, err := pgNumericToDecimal(activeOrder.Qty)
	if err != nil {
		return orderBook.Order{}, fmt.Errorf("converting qty: %w", err)
	}
	leavesQty, err := pgNumericToDecimal(activeOrder.LeavesQty)
	if err != nil {
		return orderBook.Order{}, fmt.Errorf("converting leavesQty: %w", err)
	}

	return orderBook.Order{
		ID:        activeOrder.ID,
		Price:     price,
		Qty:       qty,
		Instrument: activeOrder.Instrument,
		LeavesQty: leavesQty,
		IsBid:     activeOrder.Side == string(orderBook.Buy),
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
