package appy

var (
	mailerCtxKey = ContextKey("mailer")
)

// AttachMailer attaches the mailer to the request context.
func AttachMailer(mailer *Mailer) HandlerFunc {
	return func(c *Context) {
		c.Set(mailerCtxKey.String(), mailer)
		c.Next()
	}
}
