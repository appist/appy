package pack

import "github.com/appist/appy/mailer"

var (
	mdwMailerCtxKey = ContextKey("mailer")
)

func mdwMailer(mailer *mailer.Engine) HandlerFunc {
	return func(c *Context) {
		c.Set(mdwMailerCtxKey.String(), mailer)
		c.Next()
	}
}
