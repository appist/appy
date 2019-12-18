package http

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/appist/appy/support"
)

var (
	recoveryDunno     = []byte("???")
	recoveryCenterDot = []byte("·")
	recoveryDot       = []byte(".")
	recoverySlash     = []byte("/")
)

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery(logger *support.Logger) HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				recoveryErrorHandler(c, logger, err)
			}
		}()

		c.Next()
	}
}

func recoveryErrorHandler(c *Context, logger *support.Logger, err interface{}) {
	session := DefaultSession(c)
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
		c.Error(errors.New(e))
	case error:
		c.Error(e)
	}

	tplErrors := []template.HTML{}
	for _, err := range c.Errors {
		logger.Error(err)
		tplErrors = append(tplErrors, template.HTML(err.Error()))
	}

	headers := ""
	for key, val := range c.Request.Header {
		headers = headers + fmt.Sprintf("%s: %s<br>", key, strings.Join(val, ", "))
	}

	qsParams := ""
	for key, val := range c.Request.URL.Query() {
		qsParams = qsParams + fmt.Sprintf("%s: %s<br>", key, strings.Join(val, ", "))
	}

	if qsParams == "" {
		qsParams = "None"
	}

	c.ginHTML(http.StatusInternalServerError, "error/500", H{
		"errors":      tplErrors,
		"headers":     template.HTML(headers),
		"qsParams":    template.HTML(qsParams),
		"sessionVars": template.HTML(sessionVars),
		"title":       "500 Internal Server Error",
	})
	c.Abort()
}
