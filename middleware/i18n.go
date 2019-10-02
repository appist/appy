package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var (
	acceptLanguage = http.CanonicalHeaderKey("accept-language")

	// I18nCtxKey is the request context key for i18n.
	I18nCtxKey = "appy.i18n"
)

// I18n is a middleware that provides translations based on `Accept-Language` HTTP header.
func I18n(b *i18n.Bundle) gin.HandlerFunc {
	return func(c *gin.Context) {
		localizer := i18n.NewLocalizer(b, c.Request.Header.Get(acceptLanguage))
		c.Set(I18nCtxKey, localizer)
		c.Next()
	}
}

// I18nLocalizer returns the I18n localizer instance.
func I18nLocalizer(c *gin.Context) *i18n.Localizer {
	l, exists := c.Get(I18nCtxKey)

	if !exists {
		return nil
	}

	return l.(*i18n.Localizer)
}
