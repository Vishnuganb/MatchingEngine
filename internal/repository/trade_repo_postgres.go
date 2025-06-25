package repository

import (
	"context"

	"github.com/google/uuid"

	sqlc "MatchingEngine/internal/db/sqlc"
	"MatchingEngine/internal/model"
)

type TradeQueries interface {
	CreateTrade(ctx context.Context, params sqlc.CreateTradeParams) error
	CreateTradeSide(ctx context.Context, params sqlc.CreateTradeSideParams) error
}

type PostgresTradeRepository struct {
	queries TradeQueries
}

func NewPostgresTradeRepository(queries TradeQueries) *PostgresTradeRepository {
	return &PostgresTradeRepository{queries: queries}
}

func (r *PostgresTradeRepository) SaveTrade(ctx context.Context, trade model.TradeCaptureReport) error {
	tradeId := uuid.NewString()
	price, err := decimalToPgNumeric(trade.LastPx)
	if err != nil {
		return err
	}
	quantity, err := decimalToPgNumeric(trade.LastQty)
	if err != nil {
		return err
	}

	err = r.queries.CreateTrade(ctx, sqlc.CreateTradeParams{
		TradeReportID: tradeId,
		MsgType:       trade.MsgType,
		ExecID:        trade.ExecID,
		Symbol:        trade.Symbol,
		LastQty:       quantity,
		LastPx:        price,
		TradeDate:     trade.TradeDate,
		TransactTime:  trade.TransactTime,
	})
	if err != nil {
		return err
	}

	// Insert trade sides (552)
	for _, side := range trade.NoSides {
		err = r.queries.CreateTradeSide(ctx, sqlc.CreateTradeSideParams{
			TradeReportID: tradeId,
			Side:          mapSideToInt16(side.Side),
			OrderID:       side.OrderID,
		})
		if err != nil {
			return err
		}
	}

	return nil
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