package appy

import (
	"time"

	"github.com/hibiken/asynq"
)

// Job represents a unit of work to be performed.
type Job struct {
	*asynq.Task
}

// JobPayload holds arbitrary data needed for job processing.
type JobPayload struct {
	*asynq.Payload
}

// JobOptions specifies how a job should be processed.
type JobOptions struct {
	Deadline time.Time
	MaxRetry int
	Queue    string
	Timeout  time.Duration
}

// NewJob initializes a job instance with a unique identifier and its data for background job processing.
func NewJob(id string, data map[string]interface{}) *Job {
	return &Job{
		asynq.NewTask(id, data),
	}
}

func parseJobOptions(opts *JobOptions) []asynq.Option {
	asynqOptions := []asynq.Option{}

	if opts == nil {
		return asynqOptions
	}

	if !opts.Deadline.IsZero() {
		asynqOptions = append(asynqOptions, asynq.Deadline(opts.Deadline))
	}

	if opts.MaxRetry != 0 {
		asynqOptions = append(asynqOptions, asynq.MaxRetry(opts.MaxRetry))
	}

	if opts.Queue != "" && opts.Queue != "default" {
		asynqOptions = append(asynqOptions, asynq.Queue(opts.Queue))
	}

	if opts.Timeout != 0 {
		asynqOptions = append(asynqOptions, asynq.Timeout(opts.Timeout))
	}

	return asynqOptions
}
