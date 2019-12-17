package http

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
	return func(ctx *Context) {
		requestID := ctx.GetHeader(xRequestID)
		if requestID == "" {
			uuidV4 := uuid.NewV4()
			requestID = uuidV4.String()
		}

		ctx.Set(requestIDCtxKey.String(), requestID)
		ctx.Next()
	}
}
