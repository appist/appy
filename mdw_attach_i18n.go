package appy

import (
	"net/http"
	"strings"
)

var (
	acceptLanguage   = http.CanonicalHeaderKey("accept-language")
	i18nCtxKey       = ContextKey("i18n")
	i18nLocaleCtxKey = ContextKey("i18nLocale")
)

// AttachI18n provides translations based on `Accept-Language` HTTP header.
func AttachI18n(i18n *I18n) HandlerFunc {
	return func(c *Context) {
		languages := strings.Split(c.Request.Header.Get(acceptLanguage), ",")

		if len(languages) > 0 {
			c.Set(i18nLocaleCtxKey.String(), languages[0])
		}

		c.Set(i18nCtxKey.String(), i18n)
		c.Next()
	}
}
