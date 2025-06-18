package repository

import (
	"context"
	"fmt"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	sqlc "MatchingEngine/internal/db/sqlc"
	"MatchingEngine/internal/model"
	"MatchingEngine/orderBook"
)

type OrderQueries interface {
	CreateExecution(ctx context.Context, params sqlc.CreateExecutionParams) (sqlc.Execution, error)
}

type PostgresExecutionRepository struct {
	queries OrderQueries
}

func NewPostgresExecutionRepository(queries OrderQueries) *PostgresExecutionRepository {
	return &PostgresExecutionRepository{queries: queries}
}

func decimalToPgNumeric(d decimal.Decimal) (pgtype.Numeric, error) {
	var num pgtype.Numeric
	err := num.Scan(d.String())
	return num, err
}

func (r *PostgresExecutionRepository) SaveExecution(ctx context.Context, execReport model.ExecutionReport) (model.ExecutionReport, error) {
	executionId := uuid.NewString()
	qty, err := decimalToPgNumeric(execReport.OrderQty)
	if err != nil {
		return model.ExecutionReport{}, err
	}
	leavesQty, err := decimalToPgNumeric(execReport.LeavesQty)
	if err != nil {
		return model.ExecutionReport{}, err
	}
	price, err := decimalToPgNumeric(execReport.Price)
	if err != nil {
		return model.ExecutionReport{}, err
	}
	cumQty, err := decimalToPgNumeric(execReport.CumQty)
	if err != nil {
		return model.ExecutionReport{}, err
	}
	execution, err := r.queries.CreateExecution(ctx, sqlc.CreateExecutionParams{
		ID:          executionId,
		OrderID:     execReport.OrderID,
		OrderQty:    qty,
		LeavesQty:   leavesQty,
		Price:       price,
		Instrument:  execReport.Instrument,
		CumQty:      cumQty,
		OrderStatus: execReport.OrderStatus,
		Side:        map[bool]string{true: "buy", false: "sell"}[execReport.IsBid],
		ExecType:    execReport.ExecType,
	})
	if err != nil {
		return model.ExecutionReport{}, err
	}

	mappedOrder, err := MapExecutionToModelExecution(execution)
	if err != nil {
		return model.ExecutionReport{}, err
	}

	return mappedOrder, nil
}

func MapExecutionToModelExecution(execution sqlc.Execution) (model.ExecutionReport, error) {
	price, err := pgNumericToDecimal(execution.Price)
	if err != nil {
		return model.ExecutionReport{}, fmt.Errorf("converting price: %w", err)
	}
	qty, err := pgNumericToDecimal(execution.OrderQty)
	if err != nil {
		return model.ExecutionReport{}, fmt.Errorf("converting qty: %w", err)
	}
	leavesQty, err := pgNumericToDecimal(execution.LeavesQty)
	if err != nil {
		return model.ExecutionReport{}, fmt.Errorf("converting leavesQty: %w", err)
	}
	cumQty, err := pgNumericToDecimal(execution.CumQty)
	if err != nil {
		return model.ExecutionReport{}, fmt.Errorf("converting execQty: %w", err)
	}

	return model.ExecutionReport{
		OrderID:     execution.OrderID,
		Price:       price,
		OrderQty:    qty,
		Instrument:  execution.Instrument,
		LeavesQty:   leavesQty,
		IsBid:       execution.Side == string(orderBook.Buy),
		CumQty:      cumQty,
		OrderStatus: execution.OrderStatus,
		ExecType:    execution.ExecType,
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
