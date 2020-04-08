package pack

import (
	"net/http"
	"strings"
)

var (
	xForwardedFor = http.CanonicalHeaderKey("x-forwarded-for")
	xRealIP       = http.CanonicalHeaderKey("x-real-ip")
)

func mdwRealIP() HandlerFunc {
	return func(c *Context) {
		if rip := realIP(c.Request); rip != "" {
			c.Request.RemoteAddr = rip
		}

		c.Next()
	}
}

func realIP(r *http.Request) string {
	var ip string

	if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ", ")
		if i == -1 {
			i = len(xff)
		}
		ip = xff[:i]
	} else if xrip := r.Header.Get(xRealIP); xrip != "" {
		ip = xrip
	}

	return ip
}
