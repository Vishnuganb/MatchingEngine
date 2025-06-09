package repository

import (
	"MatchingEngine/internal/model"
	"context"

	"github.com/shopspring/decimal"
)

type OrderRepository interface {
	SaveOrder(ctx context.Context, order model.Order) (model.Order, error)
	UpdateOrder(ctx context.Context, orderID, orderStatus, execType string, leavesQty, execQty decimal.Decimal) (model.Order, error)
}