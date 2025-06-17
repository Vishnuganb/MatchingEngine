package repository

import (
	"context"

	"github.com/shopspring/decimal"

	"MatchingEngine/internal/model"
)

type OrderRepository interface {
	SaveOrder(ctx context.Context, order model.Order) (model.Order, error)
	UpdateOrder(ctx context.Context, orderID, orderStatus string, leavesQty, cumQty decimal.Decimal) (model.Order, error)
}
