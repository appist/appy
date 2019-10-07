package http

import (
	"github.com/appist/appy/middleware"
	"github.com/appist/appy/support"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
)

// ContextT is a type alias to gin.Context.
type ContextT = gin.Context

// RouterT is a type alias to gin.Engine.
type RouterT = gin.Engine

// RouterGroupT is a type alias to gin.RouterGroup.
type RouterGroupT = gin.RouterGroup

// RouteInfoT is a type alias to gin.RouteInfo.
type RouteInfoT = gin.RouteInfo

// RoutesT is a type alias to gin.IRoutes.
type RoutesT = gin.IRoutes

func newRouter(c *support.ConfigT) *gin.Engine {
	r := gin.New()
	r.AppEngine = true
	r.HandleMethodNotAllowed = true

	r.Use(middleware.CSRF(c))
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger(c))
	r.Use(middleware.RealIP())
	r.Use(middleware.ResponseHeaderFilter())
	r.Use(middleware.SessionManager(c))
	r.Use(middleware.HealthCheck(c.HTTPHealthCheckURL))
	r.Use(middleware.Prerender(c))
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(secure.New(newSecureConfig(c)))
	r.Use(middleware.Recovery())

	return r
}

func newSecureConfig(c *support.ConfigT) secure.Config {
	return secure.Config{
		IsDevelopment:           false,
		AllowedHosts:            c.HTTPAllowedHosts,
		SSLRedirect:             c.HTTPSSLRedirect,
		SSLTemporaryRedirect:    c.HTTPSSLTemporaryRedirect,
		SSLHost:                 c.HTTPSSLHost,
		STSSeconds:              c.HTTPSTSSeconds,
		STSIncludeSubdomains:    c.HTTPSTSIncludeSubdomains,
		FrameDeny:               c.HTTPFrameDeny,
		CustomFrameOptionsValue: c.HTTPCustomFrameOptionsValue,
		ContentTypeNosniff:      c.HTTPContentTypeNosniff,
		BrowserXssFilter:        c.HTTPBrowserXSSFilter,
		ContentSecurityPolicy:   c.HTTPContentSecurityPolicy,
		ReferrerPolicy:          c.HTTPReferrerPolicy,
		IENoOpen:                c.HTTPIENoOpen,
		SSLProxyHeaders:         c.HTTPSSLProxyHeaders,
	}
}
