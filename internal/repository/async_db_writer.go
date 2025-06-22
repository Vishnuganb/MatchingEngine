package repository

import (
	"context"
	"log"
	"time"
)

type AsyncDBWriterInterface interface {
	EnqueueTask(task DBTask)
}

type AsyncDBWriter struct {
	taskChannel chan DBTask
	execRepo    *PostgresExecutionRepository
	tradeRepo   *PostgresTradeRepository
	retryCount  int
	timeout     time.Duration
}

func NewAsyncDBWriter(execRepo *PostgresExecutionRepository, tradeRepo *PostgresTradeRepository, bufferSize int) *AsyncDBWriter {
	writer := &AsyncDBWriter{
		taskChannel: make(chan DBTask, bufferSize),
		execRepo:    execRepo,
		tradeRepo:   tradeRepo,
		retryCount:  3,
		timeout:     100 * time.Millisecond,
	}
	go writer.startWorkerPool(5)
	return writer
}

func (w *AsyncDBWriter) startWorkerPool(workerCount int) {
	for i := 0; i < workerCount; i++ {
		go func(workerID int) {
			for task := range w.taskChannel {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				err := w.executeTask(ctx, task)
				cancel()

				if err != nil {
					log.Printf("[Worker %d] Failed to execute task: %v", workerID, err)
				}
			}
		}(i)
	}
}

func (w *AsyncDBWriter) executeTask(ctx context.Context, task DBTask) error {
	// Type assertion to determine the repository type
	if execTask, ok := task.(SaveExecutionTask); ok {
		return execTask.Execute(ctx, w.execRepo)
	} else if tradeTask, ok := task.(SaveTradeTask); ok {
		return tradeTask.Execute(ctx, w.tradeRepo)
	}
	return nil
}

func (w *AsyncDBWriter) EnqueueTask(task DBTask) {
	for attempt := 0; attempt < w.retryCount; attempt++ {
		select {
		case w.taskChannel <- task:
			return
		case <-time.After(w.timeout):
			log.Printf("Enqueue attempt %d timed out", attempt+1)
		}
	}
	log.Println("Task channel is full after retries, dropping task")
}
