package appy

import "github.com/hibiken/asynq"

type Job struct {
	*asynq.Task
}

func NewJob(id string, data map[string]interface{}) *Job {
	return &Job{
		asynq.NewTask(id, data),
	}
}
