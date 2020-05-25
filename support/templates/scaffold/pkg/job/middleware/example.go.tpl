package middleware

import (
	"context"
	"{{.projectName}}/pkg/app"

	"github.com/appist/appy/worker"
)

// Example is a simple logging middleware that prints dummy message.
func Example(next worker.Handler) worker.Handler {
	return worker.HandlerFunc(func(ctx context.Context, job *worker.Job) error {
		app.Logger.Info("middleware example logging")
		return next.ProcessTask(ctx, job)
	})
}
