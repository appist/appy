package appy

import (
	"net/http"

	uuid "github.com/satori/go.uuid"
)

var (
	requestIDCtxKey = ContextKey("requestID")
	xRequestID      = http.CanonicalHeaderKey("x-request-id")
)

// RequestID is a middleware that injects a request ID into the context of each request.
func RequestID() HandlerFunc {
	return func(c *Context) {
		requestID := c.GetHeader(xRequestID)
		if requestID == "" {
			uuidV4 := uuid.NewV4()
			requestID = uuidV4.String()
		}

		c.Set(requestIDCtxKey.String(), requestID)
		c.Next()
	}
}
