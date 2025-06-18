package repository

import (
	"MatchingEngine/internal/model"
	"context"
)

type TradeRepository interface {
	SaveTrade(ctx context.Context, trade model.Trade) (model.Trade, error)
}
