package repository

import (
	"context"

	"github.com/google/uuid"

	sqlc "MatchingEngine/internal/db/sqlc"
	"MatchingEngine/internal/model"
)

type TradeQueries interface {
	CreateTrade(ctx context.Context, params sqlc.CreateTradeParams) (sqlc.Trade, error)
}

type PostgresTradeRepository struct {
	queries TradeQueries
}

func NewPostgresTradeRepository(queries TradeQueries) *PostgresTradeRepository {
	return &PostgresTradeRepository{queries: queries}
}

func (r *PostgresTradeRepository) SaveTrade(ctx context.Context, trade model.Trade) (model.Trade, error) {
	tradeId := uuid.NewString()
	price, err := decimalToPgNumeric(trade.Price)
	if err != nil {
		return model.Trade{}, err
	}
	quantity, err := decimalToPgNumeric(trade.Quantity)
	if err != nil {
		return model.Trade{}, err
	}

	tradeRecord, err := r.queries.CreateTrade(ctx, sqlc.CreateTradeParams{
		ID:            tradeId,
		Price:         price,
		Qty:           quantity,
		Instrument:    trade.Instrument,
		BuyerOrderID:  trade.BuyerOrderID,
		SellerOrderID: trade.SellerOrderID,
	})
	if err != nil {
		return model.Trade{}, err
	}

	mappedTrade, err := MapTradeToModelTrade(tradeRecord)
	if err != nil {
		return model.Trade{}, err
	}

	return mappedTrade, nil
}

func MapTradeToModelTrade(trade sqlc.Trade) (model.Trade, error) {
	price, err := pgNumericToDecimal(trade.Price)
	if err != nil {
		return model.Trade{}, err
	}
	quantity, err := pgNumericToDecimal(trade.Qty)
	if err != nil {
		return model.Trade{}, err
	}

	return model.Trade{
		BuyerOrderID:  trade.BuyerOrderID,
		SellerOrderID: trade.SellerOrderID,
		Price:         price,
		Quantity:      quantity,
		Instrument:    trade.Instrument,
	}, nil
}
