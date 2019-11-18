package http

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	appysupport "github.com/appist/appy/internal/support"
)

var (
	recoveryDunno     = []byte("???")
	recoveryCenterDot = []byte("·")
	recoveryDot       = []byte(".")
	recoverySlash     = []byte("/")
)

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery(logger *appysupport.Logger) HandlerFunc {
	return func(ctx *Context) {
		defer func() {
			if err := recover(); err != nil {
				recoveryErrorHandler(ctx, logger, err)
			}
		}()

		ctx.Next()
	}
}

func recoveryErrorHandler(ctx *Context, logger *appysupport.Logger, err interface{}) {
	session := DefaultSession(ctx)
	sessionVars := ""
	if session != nil && session.Values() != nil {
		for key, val := range session.Values() {
			sessionVars = sessionVars + fmt.Sprintf("%s: %+v<br>", key, val)
		}
	}

	if sessionVars == "" {
		sessionVars = "None"
	}

	switch e := err.(type) {
	case string:
		ctx.Error(errors.New(e))
	case error:
		ctx.Error(e)
	}

	tplErrors := []template.HTML{}
	for _, err := range ctx.Errors {
		logger.Error(err)
		tplErrors = append(tplErrors, template.HTML(err.Error()))
	}

	headers := ""
	for key, val := range ctx.Request.Header {
		headers = headers + fmt.Sprintf("%s: %s<br>", key, strings.Join(val, ", "))
	}

	qsParams := ""
	for key, val := range ctx.Request.URL.Query() {
		qsParams = qsParams + fmt.Sprintf("%s: %s<br>", key, strings.Join(val, ", "))
	}

	if qsParams == "" {
		qsParams = "None"
	}

	ctx.HTML(http.StatusInternalServerError, "error/500", H{
		"errors":      tplErrors,
		"headers":     template.HTML(headers),
		"qsParams":    template.HTML(qsParams),
		"sessionVars": template.HTML(sessionVars),
		"title":       "500 Internal Server Error",
	})
	ctx.Abort()
}
