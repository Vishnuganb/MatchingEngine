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
	CreateExecution(ctx context.Context, params sqlc.CreateExecutionParams) error
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

func (r *PostgresExecutionRepository) SaveExecution(ctx context.Context, execReport model.ExecutionReport) error {
	executionId := uuid.NewString()
	orderQty, err := decimalToPgNumeric(execReport.OrderQty)
	if err != nil {
		return err
	}
	leavesQty, err := decimalToPgNumeric(execReport.LeavesQty)
	if err != nil {
		return err
	}
	lastPx, err := decimalToPgNumeric(execReport.LastPx)
	if err != nil {
		return err
	}
	cumQty, err := decimalToPgNumeric(execReport.CumQty)
	if err != nil {
		return err
	}
	avgPx, err := decimalToPgNumeric(execReport.AvgPx)
	if err != nil {
		return err
	}

	err = r.queries.CreateExecution(ctx, sqlc.CreateExecutionParams{
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
		return err
	}

	return nil
}

func stringToPgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
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