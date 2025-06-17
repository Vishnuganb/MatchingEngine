package repository

import (
	"context"

	"github.com/shopspring/decimal"

	"MatchingEngine/internal/model"
)

type DBTask interface {
	Execute(ctx context.Context, repo OrderRepository) error
}

type SaveOrderTask struct {
	Order model.Order
}

func (t SaveOrderTask) Execute(ctx context.Context, repo OrderRepository) error {
	_, err := repo.SaveOrder(ctx, t.Order)
	return err
}

type UpdateOrderTask struct {
	OrderID     string
	OrderStatus string
	LeavesQty   decimal.Decimal
	CumQty      decimal.Decimal
}

func (t UpdateOrderTask) Execute(ctx context.Context, repo OrderRepository) error {
	_, err := repo.UpdateOrder(ctx, t.OrderID, t.OrderStatus, t.LeavesQty, t.CumQty)
	return err
}
