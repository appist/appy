package http

// ResponseHeaderFilter is a middleware that removes `Set-Cookie` response header when the request header
// `X-API-Only: 1` is received.
func ResponseHeaderFilter() HandlerFunc {
	return func(c *Context) {
		if IsAPIOnly(c) == true {
			c.Writer.Header().Del("Set-Cookie")
		}

		c.Next()
	}
}
