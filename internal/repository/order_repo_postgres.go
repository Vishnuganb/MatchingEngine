package repository

import (
	"MatchingEngine/internal/model"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	sqlc "MatchingEngine/internal/db/sqlc"
	"MatchingEngine/orderBook"
)

type OrderQueries interface {
	CreateActiveOrder(ctx context.Context, params sqlc.CreateActiveOrderParams) (sqlc.ActiveOrder, error)
	UpdateActiveOrder(ctx context.Context, params sqlc.UpdateActiveOrderParams) (sqlc.ActiveOrder, error)
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
	qty, err := decimalToPgNumeric(order.OrderQty)
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
		OrderQty:        qty,
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

func MapActiveOrderToOrder(activeOrder sqlc.ActiveOrder) (model.Order, error) {
	price, err := pgNumericToDecimal(activeOrder.Price)
	if err != nil {
		return model.Order{}, fmt.Errorf("converting price: %w", err)
	}
	qty, err := pgNumericToDecimal(activeOrder.OrderQty)
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
		OrderQty:        qty,
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
