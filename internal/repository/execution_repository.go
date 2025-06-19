package repository

import (
	"context"

	"MatchingEngine/internal/model"
)

type ExecutionRepository interface {
	SaveExecution(ctx context.Context, order model.ExecutionReport) (model.ExecutionReport, error)
	GetAllExecutions(ctx context.Context) ([]model.ExecutionReport, error)
}
