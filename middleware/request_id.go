package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

var requestIDCtxKey = "appy.requestID"
var requestIDHeader = http.CanonicalHeaderKey("x-request-id")

// RequestID is a middleware that injects a request ID into the context of each request.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			uuidV4 := uuid.NewV4()
			requestID = fmt.Sprintf("%s", uuidV4)
		}

		c.Set(requestIDCtxKey, requestID)
		c.Next()
	}
}
