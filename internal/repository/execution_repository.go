package repository

import (
	"MatchingEngine/internal/model"
	"context"
)

type ExecutionRepository interface {
	SaveExecution(ctx context.Context, order model.ExecutionReport) (model.ExecutionReport, error)
}
