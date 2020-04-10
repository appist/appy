package pack

import (
	"github.com/appist/appy/mailer"
	"github.com/appist/appy/support"
)

var (
	mdwMailerCtxKey = ContextKey("mailer")
)

func mdwMailer(mailer *mailer.Engine, i18n *support.I18n, server *Server) HandlerFunc {
	server.SetupMailerPreview(mailer, i18n)

	return func(c *Context) {
		c.Set(mdwMailerCtxKey.String(), mailer)
		c.Next()
	}
}
