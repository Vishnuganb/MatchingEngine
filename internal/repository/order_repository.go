package repository

import (
	"context"

	"github.com/shopspring/decimal"

	"MatchingEngine/internal/model"
)

type OrderRepository interface {
	SaveOrder(ctx context.Context, order model.Order) (model.Order, error)
	UpdateOrder(ctx context.Context, orderID, orderStatus, execType string, leavesQty, execQty decimal.Decimal) (model.Order, error)
}
