package jobs

import (
	"context"
	"tally-connector/internal/loader"
)

func importHandler(ctx context.Context, job *Job) error {
	loader.ImportAll(job.ID)
	return nil
}

var workerPool *WorkerPool

func GetDefaultWorkerPool() *WorkerPool {
	return workerPool
}

func ProcessImportQueue(ctx context.Context) {
	workerPool = NewWorkerPool(importHandler, 100)
	workerPool.AddWorker(ctx)
}
