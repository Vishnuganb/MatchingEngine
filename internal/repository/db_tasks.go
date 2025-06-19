package repository

import (
	"context"
	"fmt"

	"MatchingEngine/internal/model"
)

type DBTask interface {
	Execute(ctx context.Context, repo interface{}) error
}

type SaveExecutionTask struct {
	Execution model.ExecutionReport
}

func (t SaveExecutionTask) Execute(ctx context.Context, repo interface{}) error {
	execRepo, ok := repo.(ExecutionRepository)
	if !ok {
		return fmt.Errorf("invalid repository type")
	}
	_, err := execRepo.SaveExecution(ctx, t.Execution)
	return err
}

type GetAllExecutionsTask struct{
	Executions *[]model.ExecutionReport
}

func (t GetAllExecutionsTask) Execute(ctx context.Context, repo interface{}) error {
	execRepo, ok := repo.(ExecutionRepository)
	if !ok {
		return fmt.Errorf("invalid repository type")
	}

	executions, err := execRepo.GetAllExecutions(ctx)
	if err != nil {
		return err
	}

	*t.Executions = executions
	return nil
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
