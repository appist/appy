package core

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
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

type csrResource struct {
	assets     http.FileSystem
	fileServer http.Handler
	prefix     string
}

// AppServer is the core that serves HTTP/GRPC requests.
type AppServer struct {
	assets       http.FileSystem
	config       AppConfig
	csrResources []csrResource
	grpc         *grpc.Server
	http         *http.Server
	htmlRenderer multitemplate.Renderer
	Logger       *AppLogger
	Router       *Router
	viewHelper   template.FuncMap
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func newServer(assets http.FileSystem, c AppConfig, l *AppLogger, vh template.FuncMap) AppServer {
	// Initialize the error templates.
	renderer := multitemplate.NewRenderer()
	renderer.AddFromString("error/404", errorTpl404())
	renderer.AddFromString("error/500", errorTpl500())
	renderer.AddFromString("default/welcome", welcomeTpl())

	r := newRouter(c, l)
	r.HTMLRender = renderer

	http := &http.Server{
		Addr:              c.HTTPHost + ":" + c.HTTPPort,
		Handler:           r,
		MaxHeaderBytes:    c.HTTPMaxHeaderBytes,
		ReadTimeout:       c.HTTPReadTimeout,
		ReadHeaderTimeout: c.HTTPReadHeaderTimeout,
		WriteTimeout:      c.HTTPWriteTimeout,
		IdleTimeout:       c.HTTPIdleTimeout,
	}
	http.ErrorLog = zap.NewStdLog(l.Desugar())

	if c.HTTPSSLEnabled == true {
		http.Addr = c.HTTPHost + ":" + c.HTTPSSLPort
	}

	return AppServer{
		assets:       assets,
		config:       c,
		csrResources: []csrResource{},
		grpc:         nil,
		http:         http,
		htmlRenderer: renderer,
		Logger:       l,
		Router:       r,
		viewHelper:   vh,
	}
}

func newRouter(c AppConfig, l *AppLogger) *gin.Engine {
	r := gin.New()
	r.AppEngine = true
	r.HandleMethodNotAllowed = true

	r.Use(CSRF(c, l))
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

func newSecureConfig(c AppConfig) secure.Config {
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

// AddDefaultWelcomePage adds the default welcome page for `/` route.
func (s AppServer) AddDefaultWelcomePage() {
	routes := s.Routes()
	rootDefined := false

	for _, route := range routes {
		if route.Path == "/" {
			rootDefined = true
			break
		}
	}

	csrRootDefined := false
	if s.assets != nil {
		if _, err := s.assets.Open("/index.html"); err == nil {
			csrRootDefined = true
		}
	}

	if !rootDefined && !csrRootDefined {
		s.Router.GET("/", func(c *Context) {
			c.HTML(200, "default/welcome", nil)
		})
	}
}

// IsSSLCertsExist checks if `./tmp/ssl` exists and contains the locally trusted SSL certificates.
func (s AppServer) IsSSLCertsExist() bool {
	_, certErr := os.Stat(s.config.HTTPSSLCertPath + "/cert.pem")
	_, keyErr := os.Stat(s.config.HTTPSSLCertPath + "/key.pem")

	if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
		return false
	}

	return true
}

// Hosts returns the server hosts list.
func (s AppServer) Hosts() ([]string, error) {
	var hosts = []string{s.config.HTTPHost}

	if s.config.HTTPHost != "localhost" {
		hosts = append(hosts, "localhost")
	}

	addresses, _ := net.InterfaceAddrs()
	for _, address := range addresses {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			host := ipnet.IP.To4()
			if host != nil {
				hosts = append(hosts, host.String())
			}
		}
	}

	return hosts, nil
}

// Routes returns all the routes including those in middlewares.
func (s AppServer) Routes() []RouteInfo {
	routes := s.Router.Routes()

	if s.config.HTTPHealthCheckURL != "" {
		routes = append(routes, RouteInfo{
			Method:      "GET",
			Path:        s.config.HTTPHealthCheckURL,
			Handler:     "",
			HandlerFunc: nil,
		})
	}

	return routes
}

// PrintInfo prints the server info.
func (s AppServer) PrintInfo() {
	configPath, _, _ := getConfigInfo(s.assets)
	lines := []string{}
	lines = append(lines,
		fmt.Sprintf("* Version %s (%s), build: %s, environment: %s, config: %s",
			VERSION, runtime.Version(), Build, s.config.AppyEnv, configPath,
		),
	)

	hosts, _ := s.Hosts()
	host := fmt.Sprintf("http://%s:%s", hosts[0], s.config.HTTPPort)

	if s.config.HTTPSSLEnabled == true {
		host = fmt.Sprintf("https://%s:%s", hosts[0], s.config.HTTPSSLPort)
	}

	lines = append(lines, fmt.Sprintf("* Listening on %s", host))

	for _, line := range lines {
		if Build == "debug" {
			fmt.Println(line)
		} else {
			s.Logger.Info(line)
		}
	}
}

// InitCSR setup the client-side rendering/routing with index.html fallback.
func (s AppServer) InitCSR() {
	// Setup CSR hosting at "/".
	s.Router.Use(s.serveCSR("/", s.assets))

	// Setup CSR hosting at "/tools".
	// s.Router.Use(s.serveCSR("/tools", tools.Assets))

	s.Router.NoRoute(CSRFSkipCheck(), func(ctx *Context) {
		request := ctx.Request

		if request.Method == "GET" && !staticExtRegex.MatchString(request.URL.Path) {
			ctx.Header("Cache-Control", "no-cache")
			resource := s.csrResource(request.URL.Path)

			if resource.assets != nil {
				file, _ := resource.assets.Open("/index.html")

				if file != nil {
					data, _ := ioutil.ReadAll(file)

					if data != nil {
						ctx.Data(http.StatusOK, "text/html; charset=utf-8", data)
						return
					}
				}
			}
		}

		ctx.HTML(http.StatusNotFound, "error/404", H{
			"title": "404 Page Not Found",
		})
	})
}

func (s AppServer) csrResource(path string) csrResource {
	var (
		resource, rootResource csrResource
	)

	for _, res := range s.csrResources {
		if res.prefix == "/" {
			rootResource = res
		}

		if res.prefix != "/" && strings.HasPrefix(path, res.prefix) {
			resource = res
		}
	}

	if (csrResource{}) == resource {
		resource = rootResource
	}

	return resource
}

func (s *AppServer) serveCSR(prefix string, assets http.FileSystem) HandlerFunc {
	s.csrResources = append(s.csrResources, csrResource{
		assets:     assets,
		fileServer: http.StripPrefix(prefix, http.FileServer(assets)),
		prefix:     prefix,
	})

	return func(ctx *Context) {
		request := ctx.Request

		// Serve from the assets FS if the URL path isn't matching any of the SSR paths.
		if !isSSRPath(s.Routes(), request.URL.Path) && !strings.HasPrefix(request.URL.Path, "/"+ssrPaths["root"]) {
			resource := s.csrResource(request.URL.Path)

			if resource.assets != nil {
				requestPath := request.URL.Path
				assetPath := strings.Replace(requestPath, resource.prefix, "", 1)
				_, err := resource.assets.Open(assetPath)
				// Only serve the request from assets if the file is in the assets filesystem.
				if err == nil {
					resource.fileServer.ServeHTTP(ctx.Writer, request)
					ctx.Abort()
				}
			}
		}
	}
}

func isSSRPath(routes []RouteInfo, path string) bool {
	for _, route := range routes {
		if route.Path == path {
			return true
		}
	}

	return false
}
