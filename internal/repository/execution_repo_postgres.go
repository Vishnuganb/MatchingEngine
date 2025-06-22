package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	sqlc "MatchingEngine/internal/db/sqlc"
	"MatchingEngine/internal/model"
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
	orderQty, err := decimalToPgNumeric(execReport.OrderQty)
	if err != nil {
		return model.ExecutionReport{}, err
	}
	leavesQty, err := decimalToPgNumeric(execReport.LeavesQty)
	if err != nil {
		return model.ExecutionReport{}, err
	}
	lastPx, err := decimalToPgNumeric(execReport.LastPx)
	if err != nil {
		return model.ExecutionReport{}, err
	}
	cumQty, err := decimalToPgNumeric(execReport.CumQty)
	if err != nil {
		return model.ExecutionReport{}, err
	}
	avgPx, err := decimalToPgNumeric(execReport.AvgPx)
	if err != nil {
		return model.ExecutionReport{}, err
	}

	execution, err := r.queries.CreateExecution(ctx, sqlc.CreateExecutionParams{
		ExecID:       executionId,
		OrderID:      execReport.OrderID,
		ClOrdID:      stringToPgText(execReport.ClOrdID),
		ExecType:     string(execReport.ExecType),
		OrdStatus:    string(execReport.OrdStatus),
		Symbol:       execReport.Symbol,
		Side:         string(execReport.Side),
		OrderQty:     orderQty,
		LastShares:   decimalToPgNumericOrZero(execReport.LastShares),
		LastPx:       lastPx,
		LeavesQty:    leavesQty,
		CumQty:       cumQty,
		AvgPx:        avgPx,
		TransactTime: execReport.TransactTime,
		Text:         stringToPgText(execReport.Text),
	})
	if err != nil {
		return model.ExecutionReport{}, err
	}

	return MapExecutionToModelExecution(execution)
}

func MapExecutionToModelExecution(execution sqlc.Execution) (model.ExecutionReport, error) {
	orderQty, err := pgNumericToDecimal(execution.OrderQty)
	if err != nil {
		return model.ExecutionReport{}, fmt.Errorf("converting orderQty: %w", err)
	}
	leavesQty, err := pgNumericToDecimal(execution.LeavesQty)
	if err != nil {
		return model.ExecutionReport{}, fmt.Errorf("converting leavesQty: %w", err)
	}
	lastPx, err := pgNumericToDecimal(execution.LastPx)
	if err != nil {
		return model.ExecutionReport{}, fmt.Errorf("converting lastPx: %w", err)
	}
	cumQty, err := pgNumericToDecimal(execution.CumQty)
	if err != nil {
		return model.ExecutionReport{}, fmt.Errorf("converting cumQty: %w", err)
	}
	avgPx, err := pgNumericToDecimal(execution.AvgPx)
	if err != nil {
		return model.ExecutionReport{}, fmt.Errorf("converting avgPx: %w", err)
	}

	return model.ExecutionReport{
		ExecID:       execution.ExecID,
		OrderID:      execution.OrderID,
		ClOrdID:      pgTextToString(execution.ClOrdID),
		ExecType:     model.ExecType(execution.ExecType),
		OrdStatus:    model.OrderStatus(execution.OrdStatus),
		Symbol:       execution.Symbol,
		Side:         model.Side(execution.Side),
		OrderQty:     orderQty,
		LastShares:   pgNumericToDecimalOrZero(execution.LastShares),
		LastPx:       lastPx,
		LeavesQty:    leavesQty,
		CumQty:       cumQty,
		AvgPx:        avgPx,
		TransactTime: execution.TransactTime,
		Text:         pgTextToString(execution.Text),
	}, nil
}

func stringToPgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func pgTextToString(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
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
	return decimal.NewFromString(str)
}

func decimalToPgNumericOrZero(d decimal.Decimal) pgtype.Numeric {
	num, err := decimalToPgNumeric(d)
	if err != nil {
		return pgtype.Numeric{}
	}
	return num
}

func pgNumericToDecimalOrZero(num pgtype.Numeric) decimal.Decimal {
	dec, err := pgNumericToDecimal(num)
	if err != nil {
		return decimal.Zero
	}
	return dec
}
