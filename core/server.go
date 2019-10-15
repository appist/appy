package core

import (
	"html/template"
	"net/http"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
)

// H is a type alias to gin.H.
type H = gin.H

// HandlerFunc is a type alias to gin.HandlerFunc.
type HandlerFunc = gin.HandlerFunc

// Context is a type alias to gin.Context.
type Context = gin.Context

// Router is a type alias to gin.Engine.
type Router = gin.Engine

// RouterGroup is a type alias to gin.RouterGroup.
type RouterGroup = gin.RouterGroup

// RouteInfo is a type alias to gin.RouteInfo.
type RouteInfo = gin.RouteInfo

// Routes is a type alias to gin.IRoutes.
type Routes = gin.IRoutes

type server struct{}

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func newServer(assets http.FileSystem, viewHelper template.FuncMap) server {
	return server{}
}

func newRouter(c *AppConfig) *gin.Engine {
	r := gin.New()
	r.AppEngine = true
	r.HandleMethodNotAllowed = true

	// r.Use(middleware.CSRF(c))
	// r.Use(middleware.RequestID())
	// r.Use(middleware.RequestLogger(c))
	// r.Use(middleware.RealIP())
	// r.Use(middleware.ResponseHeaderFilter())
	// r.Use(middleware.SessionManager(c))
	// r.Use(middleware.HealthCheck(c.HTTPHealthCheckURL))
	// r.Use(middleware.Prerender(c))
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(secure.New(newSecureConfig(c)))
	// r.Use(middleware.Recovery())

	return r
}

func newSecureConfig(c *AppConfig) secure.Config {
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
