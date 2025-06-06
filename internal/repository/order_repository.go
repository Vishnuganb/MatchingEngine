package repository

import (
	"MatchingEngine/internal/model"
	"context"

	"github.com/shopspring/decimal"
)

type OrderRepository interface {
	SaveOrder(ctx context.Context, order model.Order) (model.Order, error)
	SaveEvent(ctx context.Context, event model.Event) (model.Event, error)
	UpdateOrder(ctx context.Context, orderID string, leavesQty decimal.Decimal) (model.Order, error)
	UpdateEvent(ctx context.Context, event model.Event) (model.Event, error)
}