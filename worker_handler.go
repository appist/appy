package appy

import (
	"context"
)

type WorkerHandler func(context.Context, *Job) error

func (fn WorkerHandler) ProcessTask(ctx context.Context, job *Job) error {
	return fn(ctx, job)
}
