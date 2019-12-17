package http

import (
	"net/http"
	"strings"
)

var (
	xForwardedFor = http.CanonicalHeaderKey("x-forwarded-for")
	xRealIP       = http.CanonicalHeaderKey("x-real-ip")
)

// RealIP is a middleware that sets a http.Request's RemoteAddr to the results of parsing either the X-Forwarded-For
// header or the X-Real-IP header (in that order).
func RealIP() HandlerFunc {
	return func(ctx *Context) {
		if rip := realIP(ctx.Request); rip != "" {
			ctx.Request.RemoteAddr = rip
		}

		ctx.Next()
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
