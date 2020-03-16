package appy

import "github.com/hibiken/asynq"

// Job represents a unit of work to be performed.
type Job struct {
	*asynq.Task
}

// NewJob initializes a job instance with a unique identifier and its data for background job processing.
func NewJob(id string, data map[string]interface{}) *Job {
	return &Job{
		asynq.NewTask(id, data),
	}
}
