package repository

import (
	"context"

	"MatchingEngine/internal/model"
)

type ExecutionRepository interface {
	SaveExecution(ctx context.Context, order model.ExecutionReport) error
}
