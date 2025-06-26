package service

import (
	"MatchingEngine/internal/model"
	"MatchingEngine/internal/repository"
)

type ExecutionService struct {
	asyncWriter repository.AsyncDBWriterInterface
	Notifier Notifier
}

func NewExecutionService(asyncWriter repository.AsyncDBWriterInterface, notifier Notifier) *ExecutionService {
	return &ExecutionService{
		asyncWriter: asyncWriter,
		Notifier: notifier,
	}
}

func (e *ExecutionService) SaveExecutionAsync(execution model.ExecutionReport) {
	e.asyncWriter.EnqueueTask(repository.SaveExecutionTask{
		Execution: execution,
	})
}
