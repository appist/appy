package pack

func mdwAPIOnly() HandlerFunc {
	return func(c *Context) {
		if c.IsAPIOnly() {
			c.Writer.Header().Del("Set-Cookie")
		}

		c.Next()
	}
}
