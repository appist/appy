package http

import (
	"net/http"
	"strings"

	"github.com/appist/appy/support"
)

var (
	acceptLanguage   = http.CanonicalHeaderKey("accept-language")
	i18nLocaleCtxKey = ContextKey("i18nLocale")
)

// I18n provides translations based on `Accept-Language` HTTP header.
func I18n(i18n *support.I18n) HandlerFunc {
	return func(c *Context) {
		languages := strings.Split(c.Request.Header.Get(acceptLanguage), ",")

		if len(languages) > 0 {
			c.Set(i18nLocaleCtxKey.String(), languages[0])
		}

		c.i18n = i18n
		c.Next()
	}
}
