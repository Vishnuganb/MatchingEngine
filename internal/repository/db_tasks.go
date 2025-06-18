package repository

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	"MatchingEngine/internal/model"
)

type DBTask interface {
	Execute(ctx context.Context, repo interface{}) error
}

type SaveOrderTask struct {
	Order model.Order
}

func (t SaveOrderTask) Execute(ctx context.Context, repo interface{}) error {
	orderRepo, ok := repo.(OrderRepository)
	if !ok {
		return fmt.Errorf("invalid repository type")
	}
	_, err := orderRepo.SaveOrder(ctx, t.Order)
	return err
}

type UpdateOrderTask struct {
	OrderID     string
	OrderStatus string
	LeavesQty   decimal.Decimal
	CumQty      decimal.Decimal
	Price       decimal.Decimal
}

func (t UpdateOrderTask) Execute(ctx context.Context, repo interface{}) error {
	orderRepo, ok := repo.(OrderRepository)
	if !ok {
		return fmt.Errorf("invalid repository type")
	}
	_, err := orderRepo.UpdateOrder(ctx, t.OrderID, t.OrderStatus, t.LeavesQty, t.CumQty, t.Price)
	return err
}

type SaveTradeTask struct {
	Trade model.Trade
}

func (t SaveTradeTask) Execute(ctx context.Context, repo interface{}) error {
	tradeRepo, ok := repo.(TradeRepository)
	if !ok {
		return fmt.Errorf("invalid repository type")
	}
	_, err := tradeRepo.SaveTrade(ctx, t.Trade)
	return err
}
