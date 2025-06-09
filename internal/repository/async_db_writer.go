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
	repo        *PostgresOrderRepository
	retryCount  int
	timeout     time.Duration
}

func NewAsyncDBWriter(repo *PostgresOrderRepository, bufferSize int) *AsyncDBWriter {
	writer := &AsyncDBWriter{
		taskChannel: make(chan DBTask, bufferSize),
		repo:        repo,
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
				err := task.Execute(ctx, w.repo)
				cancel()

				if err != nil {
					log.Printf("[Worker %d] Failed to execute task: %v", workerID, err)
				}
			}
		}(i)
	}
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
