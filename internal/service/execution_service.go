package service

import (
	"MatchingEngine/internal/model"
	"MatchingEngine/internal/repository"
)

type ExecutionService struct {
	asyncWriter repository.AsyncDBWriterInterface
}

func NewExecutionService(asyncWriter repository.AsyncDBWriterInterface) *ExecutionService {
	return &ExecutionService{
		asyncWriter: asyncWriter,
	}
}

func (e *ExecutionService) SaveExecutionAsync(execution model.ExecutionReport) {
	e.asyncWriter.EnqueueTask(repository.SaveExecutionTask{
		Execution: execution,
	})
}
