package middleware

import (
	"{{.Project.Name}}/pkg/app"

	"github.com/appist/appy"
)

// Example is a simple logging middleware that prints dummy message.
func Example() appy.HandlerFunc {
	return func(c *appy.Context) {
		app.Logger.Info("middleware example logging")
		c.Next()
	}
}
