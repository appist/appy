package http

import (
	"net/http"

	am "github.com/appist/appy/mailer"
	"github.com/appist/appy/support"
)

var (
	mailerCtxKey = ContextKey("mailer")
)

// Mailer provides email sending via SMTP protocol.
func Mailer(i18n *support.I18n, mailer *am.Mailer, server *Server) HandlerFunc {
	if support.IsDebugBuild() {
		setupPreview(i18n, mailer, server)
	}

	return func(c *Context) {
		c.Set(mailerCtxKey.String(), mailer)
		c.Next()
	}
}

func setupPreview(i18n *support.I18n, mailer *am.Mailer, server *Server) {
	server.HTMLRenderer().AddFromString("mailer/preview", am.PreviewTpl())

	// Serve the preview listing page.
	server.GET(server.Config().MailerPreviewBaseURL, func(c *Context) {
		name := c.DefaultQuery("name", "")
		if name == "" && len(mailer.Previews()) > 0 {
			for _, preview := range mailer.Previews() {
				name = preview.Template
				break
			}
		}

		locale := c.DefaultQuery("locale", server.Config().I18nDefaultLocale)
		preview := mailer.Previews()[name]
		preview.Locale = locale

		subject := i18n.T(preview.Subject, preview.Locale)
		if subject != "" {
			preview.Subject = subject
		}

		c.DefaultHTML(http.StatusOK, "mailer/preview", support.H{
			"baseURL":  server.Config().MailerPreviewBaseURL,
			"previews": mailer.Previews(),
			"title":    "Mailer Preview",
			"name":     name,
			"ext":      c.DefaultQuery("ext", "html"),
			"locale":   locale,
			"locales":  i18n.Locales(),
			"mail":     preview,
		})
	})

	// Serve the preview content.
	server.GET(server.Config().MailerPreviewBaseURL+"/preview", func(c *Context) {
		name := c.Query("name")
		preview, exists := mailer.Previews()[name]
		if !exists {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		var (
			contentType string
			content     []byte
		)

		preview.Locale = c.DefaultQuery("locale", server.Config().I18nDefaultLocale)
		ext := c.DefaultQuery("ext", "html")
		switch ext {
		case "html":
			contentType = "text/html"
			email, err := mailer.ComposeEmail(preview)
			if err != nil {
				panic(err)
			}

			content = email.HTML
		case "txt":
			contentType = "text/plain"
			email, err := mailer.ComposeEmail(preview)

			if err != nil {
				panic(err)
			}

			content = email.Text
		}

		c.Writer.Header().Del(http.CanonicalHeaderKey("x-frame-options"))
		c.Data(http.StatusOK, contentType, content)
	})
}
