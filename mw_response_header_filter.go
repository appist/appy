package appy

// ResponseHeaderFilter is a middleware that removes `Set-Cookie` response header when the request header
// `X-API-Only: 1` is received.
func ResponseHeaderFilter() HandlerFunc {
	return func(ctx *Context) {
		if IsAPIOnly(ctx) == true {
			ctx.Writer.Header().Del("Set-Cookie")
		}

		ctx.Next()
	}
}
