package appy

import (
	"fmt"
	"net/http"

	uuid "github.com/satori/go.uuid"
)

var (
	requestIDCtxKey = "appy.requestID"
	xRequestID      = http.CanonicalHeaderKey("x-request-id")
)

// RequestID is a middleware that injects a request ID into the context of each request.
func RequestID() HandlerFunc {
	return func(ctx *Context) {
		requestID := ctx.GetHeader(xRequestID)
		if requestID == "" {
			uuidV4 := uuid.NewV4()
			requestID = fmt.Sprintf("%s", uuidV4)
		}

		ctx.Set(requestIDCtxKey, requestID)
		ctx.Next()
	}
}
