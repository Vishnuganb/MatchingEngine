package repository

import (
	"MatchingEngine/internal/model"
	"context"
	"github.com/shopspring/decimal"
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
	ExecType    string
	LeavesQty   decimal.Decimal
	ExecQty     decimal.Decimal
}

func (t UpdateOrderTask) Execute(ctx context.Context, repo OrderRepository) error {
	_, err := repo.UpdateOrder(ctx, t.OrderID, t.OrderStatus, t.ExecType, t.LeavesQty, t.ExecQty)
	return err
}
