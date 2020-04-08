package pack

import (
	"net/http"
	"strings"

	"github.com/appist/appy/support"
)

var (
	acceptLanguage      = http.CanonicalHeaderKey("accept-language")
	mdwI18nCtxKey       = ContextKey("mdwI18n")
	mdwI18nLocaleCtxKey = ContextKey("mdwI18nLocale")
)

func mdwI18n(i18n *support.I18n) HandlerFunc {
	return func(c *Context) {
		languages := strings.Split(c.Request.Header.Get(acceptLanguage), ",")

		if len(languages) > 0 {
			c.Set(mdwI18nLocaleCtxKey.String(), languages[0])
		}

		c.Set(mdwI18nCtxKey.String(), i18n)
		c.Next()
	}
}
