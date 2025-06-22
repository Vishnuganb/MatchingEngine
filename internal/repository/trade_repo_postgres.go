package repository

import (
	"context"

	"github.com/google/uuid"

	sqlc "MatchingEngine/internal/db/sqlc"
	"MatchingEngine/internal/model"
)

type TradeQueries interface {
	CreateTrade(ctx context.Context, params sqlc.CreateTradeParams) (sqlc.TradeCaptureReport, error)
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
		TradeReportID:      tradeId,
		ExecID:             trade.ExecID,
		OrderID:            trade.OrderID,
		ClOrdID:            stringToPgText(trade.ClOrdID),
		Symbol:             trade.Symbol,
		Side:               string(trade.Side),
		LastQty:            quantity,
		LastPx:             price,
		TradeDate:          trade.TradeDate,
		TransactTime:       trade.TransactTime,
		PreviouslyReported: trade.PreviouslyReported,
		Text:               stringToPgText(trade.Text),
	})
	if err != nil {
		return model.TradeCaptureReport{}, err
	}

	return MapTradeToModelTrade(tradeRecord)
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
		TradeReportID:      trade.TradeReportID,
		ExecID:             trade.ExecID,
		OrderID:            trade.OrderID,
		ClOrdID:            pgTextToString(trade.ClOrdID),
		Symbol:             trade.Symbol,
		Side:               model.Side(trade.Side),
		LastQty:            quantity,
		LastPx:             price,
		TradeDate:          trade.TradeDate,
		TransactTime:       trade.TransactTime,
		PreviouslyReported: trade.PreviouslyReported,
		Text:               pgTextToString(trade.Text),
	}, nil
}
