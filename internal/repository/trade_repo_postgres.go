package repository

import (
	"context"

	"github.com/google/uuid"

	sqlc "MatchingEngine/internal/db/sqlc"
	"MatchingEngine/internal/model"
)

type TradeQueries interface {
	CreateTrade(ctx context.Context, params sqlc.CreateTradeParams) (sqlc.TradeCaptureReport, error)
	CreateTradeSide(ctx context.Context, params sqlc.CreateTradeSideParams) (sqlc.TradeSide, error)
}

type PostgresTradeRepository struct {
	queries TradeQueries
}

func NewPostgresTradeRepository(queries TradeQueries) *PostgresTradeRepository {
	return &PostgresTradeRepository{queries: queries}
}

func (r *PostgresTradeRepository) SaveTrade(ctx context.Context, trade model.TradeCaptureReport) (model.TradeCaptureReport, error) {
	tradeId := uuid.NewString()
	price, err := decimalToPgNumeric(trade.LastPx)
	if err != nil {
		return model.TradeCaptureReport{}, err
	}
	quantity, err := decimalToPgNumeric(trade.LastQty)
	if err != nil {
		return model.TradeCaptureReport{}, err
	}

	tradeRecord, err := r.queries.CreateTrade(ctx, sqlc.CreateTradeParams{
		TradeReportID: tradeId,
		ExecID:        trade.ExecID,
		Symbol:        trade.Symbol,
		LastQty:       quantity,
		LastPx:        price,
		TradeDate:     trade.TradeDate,
		TransactTime:  trade.TransactTime,
	})
	if err != nil {
		return model.TradeCaptureReport{}, err
	}

	// Insert trade sides (552)
	for _, side := range trade.NoSides {
		_, err := r.queries.CreateTradeSide(ctx, sqlc.CreateTradeSideParams{
			TradeReportID: tradeId,
			Side:          mapSideToInt16(side.Side),
			OrderID:       side.OrderID,
		})
		if err != nil {
			return model.TradeCaptureReport{}, err
		}
	}

	return MapTradeToModelTrade(tradeRecord)
}

func mapSideToInt16(side model.Side) int16 {
	switch side {
	case model.Buy:
		return 1
	case model.Sell:
		return 2
	default:
		return 0
	}
}

func MapTradeToModelTrade(trade sqlc.TradeCaptureReport) (model.TradeCaptureReport, error) {
	price, err := pgNumericToDecimal(trade.LastPx)
	if err != nil {
		return model.TradeCaptureReport{}, err
	}
	quantity, err := pgNumericToDecimal(trade.LastQty)
	if err != nil {
		return model.TradeCaptureReport{}, err
	}

	return model.TradeCaptureReport{
		TradeReportID: trade.TradeReportID,
		ExecID:        trade.ExecID,
		Symbol:        trade.Symbol,
		LastQty:       quantity,
		LastPx:        price,
		TradeDate:     trade.TradeDate,
		TransactTime:  trade.TransactTime,
	}, nil
}
