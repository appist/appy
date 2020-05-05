package pack

import (
	"net/http"

	uuid "github.com/gofrs/uuid"
)

var (
	mdwReqIDCtxKey = ContextKey("mdwReqID")
	xReqID         = http.CanonicalHeaderKey("x-request-id")
)

func mdwReqID() HandlerFunc {
	return func(c *Context) {
		reqID := c.GetHeader(xReqID)
		if reqID == "" {
			uuidV4, _ := uuid.NewV4()
			reqID = uuidV4.String()
		}

		c.Set(mdwReqIDCtxKey.String(), reqID)
		c.Next()
	}
}
