package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/appist/appy/support"
)

// RequestLogger is a middleware that logs the start and end of each request, along with some useful data about what
// was requested, what the response status was, and how long it took to return.
func RequestLogger(config *support.Config, logger *support.Logger) HandlerFunc {
	return func(c *Context) {
		requestID, _ := c.Get(requestIDCtxKey.String())
		start := time.Now()
		c.Next()

		r := c.Request
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}

		logger.Infof("[HTTP] %s %s '%s://%s%s %s' from %s - %d %dB in %s", requestID, r.Method, scheme, r.Host, filterParams(r, config),
			r.Proto, r.RemoteAddr, c.Writer.Status(), c.Writer.Size(), time.Since(start))
	}
}

func filterParams(r *http.Request, config *support.Config) string {
	var queryParams []string
	splits := strings.Split(r.RequestURI, "?")
	baseURI := splits[0]

	for key, value := range r.URL.Query() {
		needsFilter := false

		for _, filter := range config.HTTPLogFilterParameters {
			if strings.Contains(key, filter) == true {
				needsFilter = true
				break
			}
		}

		newValue := value[0]
		if needsFilter == true {
			newValue = "[FILTERED]"
		}

		queryParams = append(queryParams, key+"="+newValue)
	}

	if len(queryParams) == 0 {
		return baseURI
	}

	return baseURI + "?" + strings.Join(queryParams, "&")
}
