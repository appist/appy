package middleware

import (
	"context"
	"{{.projectName}}/pkg/app"

	"github.com/appist/appy"
)

// Example is a simple logging middleware that prints dummy message.
func Example(next appy.WorkerHandler) appy.WorkerHandler {
	return appy.WorkerHandlerFunc(func(ctx context.Context, job *appy.Job) error {
		app.Logger.Info("middleware example logging")
		return next.ProcessTask(ctx, job)
	})
}
