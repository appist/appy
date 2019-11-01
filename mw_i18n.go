package appy

import (
	"net/http"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var (
	acceptLanguage   = http.CanonicalHeaderKey("accept-language")
	i18nCtxKey       = "appy.i18n"
	i18nLocaleCtxKey = "appy.i18nLocale"
)

// I18n is a middleware that provides translations based on `Accept-Language` HTTP header.
func I18n(b *i18n.Bundle) HandlerFunc {
	return func(ctx *Context) {
		languages := strings.Split(ctx.Request.Header.Get(acceptLanguage), ",")
		localizer := i18n.NewLocalizer(b, languages...)
		ctx.Set(i18nCtxKey, localizer)

		if len(languages) > 0 {
			ctx.Set(i18nLocaleCtxKey, languages[0])
		}

		ctx.Next()
	}
}

// I18nLocalizer returns the I18n localizer instance.
func I18nLocalizer(ctx *Context) *i18n.Localizer {
	localizer, exists := ctx.Get(i18nCtxKey)

	if !exists {
		return nil
	}

	return localizer.(*i18n.Localizer)
}

// I18nLocale returns the I18n locale.
func I18nLocale(ctx *Context) string {
	locale, exists := ctx.Get(i18nLocaleCtxKey)

	if locale == "" || !exists {
		return "en"
	}

	return locale.(string)
}

// T translates a message based on the given key. Furthermore, we can pass in template data with `Count` in it to
// support singular/plural cases.
func T(ctx *Context, key string, args ...map[string]interface{}) string {
	localizer := I18nLocalizer(ctx)

	if len(args) < 1 {
		msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: key})

		if err != nil {
			return ""
		}

		return msg
	}

	count := -1
	if _, ok := args[0]["Count"]; ok {
		count = args[0]["Count"].(int)
	}

	countKey := key
	if count != -1 {
		switch count {
		case 0:
			countKey = key + ".Zero"
		case 1:
			countKey = key + ".One"
		default:
			countKey = key + ".Other"
		}
	}

	msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: countKey, TemplateData: args[0]})

	if err != nil {
		return ""
	}

	return msg
}
