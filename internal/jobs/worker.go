package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"tally-connector/ws"
)

type Job struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Payload any    `json:"payload"`
}

// JobHandler defines the function signature for processing jobs.
type JobHandler func(ctx context.Context, job *Job) error

type WorkerPool struct {
	jobStatus  sync.Map
	jobs       chan *Job
	jobHandler JobHandler
}

// NewWorkerPool creates and initializes a new WorkerPool with a buffered channel.
func NewWorkerPool(jobHandler JobHandler, bufferSize int) *WorkerPool {
	return &WorkerPool{
		jobHandler: jobHandler,
		jobs:       make(chan *Job, bufferSize),
	}
}

// AddWorker starts a worker goroutine that processes jobs from the channel.
func (wp *WorkerPool) AddWorker(ctx context.Context) {
	go func() {
		log.Println("Worker started. Waiting for jobs...")

		for {
			select {
			case <-ctx.Done():
				log.Println("Worker shutting down...")
				return
			case job := <-wp.jobs:
				job.Status = "processing"
				wp.jobStatus.Store(job.ID, job)
				log.Printf("Processing job: %s\n", job.ID)

				wp.jobHandler(ctx, job)
				wp.jobStatus.Delete(job.ID)

				hub := ws.GetHub("job-queue")

				hub.Broadcast(ws.NewMessage("job_completed", fmt.Sprintf("Job %s completed", job.ID)))
				hub.Broadcast(ws.NewMessage("job_list", wp.GetJobsWithDetails()))
			}
		}
	}()
}

// AddJob adds a new job to the channel and tracks its status.
func (wp *WorkerPool) AddJob(job *Job) error {
	select {
	case wp.jobs <- job:
		job.Status = "queued"

		if _, ok := wp.jobStatus.Load(job.ID); ok {
			log.Printf("Job with ID %s is already in the queue.\n", job.ID)
			return fmt.Errorf("job with ID %s is already in the queue", job.ID)
		}

		wp.jobStatus.Store(job.ID, job)
		log.Printf("Added job with ID %s to queue.\n", job.ID)

		hub := ws.GetHub("job-queue")
		hub.Broadcast(ws.NewMessage("job_added", fmt.Sprintf("Job %s added to the queue", job.ID)))

		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}

// GetJobs returns a snapshot of all tracked jobs.
func (wp *WorkerPool) GetJobs() []*Job {
	var jobs = []*Job{}

	log.Println("Fetching all jobs from worker pool...")

	wp.jobStatus.Range(func(key, value any) bool {

		if job, ok := value.(*Job); ok {
			jobs = append(jobs, job)
		}

		return true
	})

	return jobs
}

type JobDetails struct {
	InQueue    int    `json:"inqueue"`
	Processing int    `json:"processing"`
	Jobs       []*Job `json:"jobs"`
}

func (wp *WorkerPool) GetJobsWithDetails() JobDetails {
	jobs := wp.GetJobs()

	var inqueue int
	var processing int

	for _, job := range jobs {
		switch job.Status {
		case "queued":
			inqueue++
		case "processing":
			processing++
		}
	}

	return JobDetails{
		InQueue:    inqueue,
		Processing: processing,
		Jobs:       jobs,
	}
}

func (wp *WorkerPool) DeleteJob(jobID string) {
	wp.jobStatus.Delete(jobID)
}
