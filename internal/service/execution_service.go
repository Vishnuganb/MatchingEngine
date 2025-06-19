package service

import (
	"MatchingEngine/internal/model"
	"MatchingEngine/internal/repository"
	"fmt"
)

type ExecutionService struct {
	asyncWriter repository.AsyncDBWriterInterface
	syncReader  repository.SyncDBReader
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

func (e *ExecutionService) GetAllExecutions() ([]model.ExecutionReport, error) {
	executions, err := e.syncReader.DequeueTasks(repository.GetAllExecutionsTask{})
	if err != nil {
		return nil, fmt.Errorf("failed to get all executions: %w", err) // note the error handling needs to change
	}
	return executions, nil
}
