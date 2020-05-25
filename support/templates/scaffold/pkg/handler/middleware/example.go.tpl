package middleware

import (
	"{{.projectName}}/pkg/app"

	"github.com/appist/appy/pack"
)

// Example is a simple logging middleware that prints dummy message.
func Example() pack.HandlerFunc {
	return func(c *pack.Context) {
		app.Logger.Info("middleware example logging")
		c.Next()
	}
}
