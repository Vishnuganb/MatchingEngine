package repository

import (
	"context"

	"MatchingEngine/internal/model"
)

type TradeRepository interface {
	SaveTrade(ctx context.Context, trade model.TradeCaptureReport) (model.TradeCaptureReport, error)
}
