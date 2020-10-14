package worker

import (
	"time"

	"github.com/hibiken/asynq"
)

type (
	// Job represents a unit of work to be performed.
	Job = asynq.Task

	// JobResult holds enqueued job's metadata.
	JobResult = asynq.Result

	// JobPayload holds arbitrary data needed for job processing.
	JobPayload = asynq.Payload

	// JobOptions specifies how a job should be processed.
	JobOptions struct {
		Deadline  time.Time
		MaxRetry  int
		ProcessAt time.Time
		ProcessIn time.Duration
		Queue     string
		Timeout   time.Duration
		UniqueTTL time.Duration
	}
)

// NewJob initializes a job with a unique identifier and its data for background job processing.
func NewJob(id string, data map[string]interface{}) *Job {
	return asynq.NewTask(id, data)
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

	if !opts.ProcessAt.IsZero() {
		asynqOptions = append(asynqOptions, asynq.ProcessAt(opts.ProcessAt))
	}

	if opts.ProcessIn != 0 {
		asynqOptions = append(asynqOptions, asynq.ProcessIn(opts.ProcessIn))
	}

	if opts.Queue != "" && opts.Queue != "default" {
		asynqOptions = append(asynqOptions, asynq.Queue(opts.Queue))
	}

	if opts.Timeout != 0 {
		asynqOptions = append(asynqOptions, asynq.Timeout(opts.Timeout))
	}

	if opts.UniqueTTL != 0 {
		asynqOptions = append(asynqOptions, asynq.Unique(opts.UniqueTTL))
	}

	return asynqOptions
}
