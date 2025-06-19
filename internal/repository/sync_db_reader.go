package repository

import (
	"context"
	"log"
	"time"

	"MatchingEngine/internal/model"
)

type SyncDBReaderInteface interface {
	DequeueTasks(task DBTask) ([]model.ExecutionReport, error)
}

type SyncDBReader struct {
	executionRepo ExecutionRepository
	retryCount    int
	timeout       time.Duration
}

func NewSyncDBReader(executionRepo *PostgresExecutionRepository) *SyncDBReader {
	return &SyncDBReader{
		executionRepo: executionRepo,
		retryCount:    3,
		timeout:       10 * time.Second,
	}
}

func (r *SyncDBReader) DequeueTasks(task DBTask) ([]model.ExecutionReport, error) {
	execTask, ok := task.(*GetAllExecutionsTask) // Correct type assertion
	if !ok {
		log.Println("Invalid task type")
		return nil, nil // or an appropriate error
	}
	err := execTask.Execute(context.Background(), r.executionRepo)
	if err != nil {
		return nil, err
	}
	return *execTask.Executions, nil
}
