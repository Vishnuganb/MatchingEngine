package repository

import (
	"MatchingEngine/internal/model"
	"context"
	"github.com/shopspring/decimal"
)

type DBTask interface {
	Execute(ctx context.Context, repo OrderRepository) error
}

type SaveEventTask struct {
	Event model.Event
}

func (t SaveEventTask) Execute(ctx context.Context, repo OrderRepository) error {
	_, err := repo.SaveEvent(ctx, t.Event)
	return err
}

type SaveOrderTask struct {
	Order model.Order
}

func (t SaveOrderTask) Execute(ctx context.Context, repo OrderRepository) error {
	_, err := repo.SaveOrder(ctx, t.Order)
	return err
}

type UpdateOrderTask struct {
	OrderID    string
	LeavesQty  decimal.Decimal
}

func (t UpdateOrderTask) Execute(ctx context.Context, repo OrderRepository) error {
	_, err := repo.UpdateOrder(ctx, t.OrderID, t.LeavesQty)
	return err
}

type CancelEventTask struct {
	Event model.Event
}

func (t CancelEventTask) Execute(ctx context.Context, repo OrderRepository) error {
	_, err := repo.UpdateEvent(ctx, t.Event)
	return err
}
