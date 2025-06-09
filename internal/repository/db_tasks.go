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
	OrderID   string
	LeavesQty decimal.Decimal
}

func (t UpdateOrderTask) Execute(ctx context.Context, repo OrderRepository) error {
	_, err := repo.UpdateOrder(ctx, t.OrderID, t.LeavesQty)
	return err
}

type CancelOrderTask struct {
	OrderID     string
	OrderStatus string
	ExecType    string
}

func (t CancelOrderTask) Execute(ctx context.Context, repo OrderRepository) error {
	_, err := repo.UpdateOrder(ctx, t.OrderID, t.OrderStatus, t.ExecType)
	return err
}
