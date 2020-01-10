package appy

// APIOnlyResponse is a middleware that removes `Set-Cookie` response header when the request header
// `X-API-Only: 1` is received.
func APIOnlyResponse() HandlerFunc {
	return func(c *Context) {
		if c.IsAPIOnly() {
			c.Writer.Header().Del("Set-Cookie")
		}

		c.Next()
	}
}
