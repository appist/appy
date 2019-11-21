package http

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	gqlgenHandler "github.com/99designs/gqlgen/handler"
	"github.com/BurntSushi/toml"
	appysupport "github.com/appist/appy/internal/support"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/vektah/gqlparser/gqlerror"
	"go.uber.org/zap"
	"golang.org/x/text/language"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

type (
	// Server is the engine that serves HTTP requests.
	Server struct {
		assets       http.FileSystem
		config       *appysupport.Config
		csrResources []csrResource
		grpc         *grpc.Server
		http         *http.Server
		https        *http.Server
		htmlRenderer multitemplate.Renderer
		i18nBundle   *i18n.Bundle
		logger       *appysupport.Logger
		router       *Router
		csrPaths     map[string]string
		ssrPaths     map[string]string
		viewHelper   template.FuncMap
	}

	csrResource struct {
		assets     http.FileSystem
		fileServer http.Handler
		prefix     string
	}
)

var (
	reservedViewDirs = []string{"layouts", "includes"}
	_staticExtRegex  = regexp.MustCompile(`\.(bmp|css|csv|eot|exif|gif|html|ico|ini|jpg|jpeg|js|json|mp4|otf|pdf|png|svg|webp|woff|woff2|tiff|ttf|toml|txt|xml|xlsx|yml|yaml)$`)
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

// NewServer initializes Server instance.
func NewServer(c *appysupport.Config, l *appysupport.Logger, assets http.FileSystem, viewHelper template.FuncMap) *Server {
	// Initialize the error templates.
	renderer := multitemplate.NewRenderer()
	renderer.AddFromString("error/404", errorTpl404())
	renderer.AddFromString("error/500", errorTpl500())
	renderer.AddFromString("default/welcome", welcomeTpl())

	r := newRouter(c, l)
	r.HTMLRender = renderer

	h := &http.Server{
		Addr:              c.HTTPHost + ":" + c.HTTPPort,
		Handler:           r,
		MaxHeaderBytes:    c.HTTPMaxHeaderBytes,
		ReadTimeout:       c.HTTPReadTimeout,
		ReadHeaderTimeout: c.HTTPReadHeaderTimeout,
		WriteTimeout:      c.HTTPWriteTimeout,
		IdleTimeout:       c.HTTPIdleTimeout,
	}
	h.ErrorLog = zap.NewStdLog(l.Desugar())

	hs := &http.Server{
		Addr:              c.HTTPHost + ":" + c.HTTPSSLPort,
		Handler:           r,
		MaxHeaderBytes:    c.HTTPMaxHeaderBytes,
		ReadTimeout:       c.HTTPReadTimeout,
		ReadHeaderTimeout: c.HTTPReadHeaderTimeout,
		WriteTimeout:      c.HTTPWriteTimeout,
		IdleTimeout:       c.HTTPIdleTimeout,
	}
	hs.ErrorLog = zap.NewStdLog(l.Desugar())

	return &Server{
		assets:       assets,
		config:       c,
		csrResources: []csrResource{},
		grpc:         nil,
		http:         h,
		https:        hs,
		htmlRenderer: renderer,
		logger:       l,
		router:       r,
		csrPaths:     appysupport.CSRPaths,
		ssrPaths:     appysupport.SSRPaths,
		viewHelper:   viewHelper,
	}
}

func newRouter(c *appysupport.Config, l *appysupport.Logger) *Router {
	r := gin.New()
	r.AppEngine = true
	r.HandleMethodNotAllowed = true

	r.Use(CSRF(c, l))
	r.Use(RequestID())
	r.Use(RequestLogger(c, l))
	r.Use(RealIP())
	r.Use(ResponseHeaderFilter())
	r.Use(SessionManager(c))
	r.Use(HealthCheck(c.HTTPHealthCheckURL))
	r.Use(Prerender(c, l))
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(secure.New(newSecureConfig(c)))
	r.Use(Recovery(l))

	return r
}

func newSecureConfig(c *appysupport.Config) secure.Config {
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

// Hosts returns the server hosts list.
func (s Server) Hosts() ([]string, error) {
	var hosts = []string{}

	if s.config.HTTPHost != "" && !appysupport.ArrayContains(hosts, s.config.HTTPHost) {
		hosts = append(hosts, s.config.HTTPHost)
	}

	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

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

// Config returns the server's configuration.
func (s Server) Config() *appysupport.Config {
	return s.config
}

// HTMLRenderer returns the server's HTML renderer.
func (s Server) HTMLRenderer() multitemplate.Renderer {
	return s.htmlRenderer
}

// HTTP returns the HTTP server instance.
func (s Server) HTTP() *http.Server {
	return s.http
}

// HTTPS returns the HTTPS server instance.
func (s Server) HTTPS() *http.Server {
	return s.https
}

// Middleware returns the server's middleware list.
func (s Server) Middleware() Middleware {
	return s.router.Handlers
}

// IsSSLCertExisted checks if `./tmp/ssl` exists and contains the locally trusted SSL certificates.
func (s Server) IsSSLCertExisted() bool {
	_, certErr := os.Stat(s.config.HTTPSSLCertPath + "/cert.pem")
	_, keyErr := os.Stat(s.config.HTTPSSLCertPath + "/key.pem")

	if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
		return false
	}

	return true
}

// Info returns the server info.
func (s Server) Info() []string {
	lines := []string{}
	lines = append(lines,
		fmt.Sprintf("* appy %s (%s), build: %s, environment: %s, config: %s",
			appysupport.VERSION, runtime.Version(), appysupport.Build, s.config.AppyEnv, s.config.Path,
		),
	)

	hosts, _ := s.Hosts()
	host := fmt.Sprintf("http://%s:%s", hosts[0], s.config.HTTPPort)

	if s.config.HTTPSSLEnabled == true {
		host += fmt.Sprintf(", https://%s:%s", hosts[0], s.config.HTTPSSLPort)
	}

	return append(lines, fmt.Sprintf("* Listening on %s", host))
}

// Routes returns all the routes including those in middlewares.
func (s Server) Routes() []RouteInfo {
	routes := s.router.Routes()

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

// SetupGraphQL sets up the GraphQL stack.
func (s Server) SetupGraphQL(path string, schema graphql.ExecutableSchema, opts []gqlgenHandler.Option, middleware ...HandlerFunc) {
	if opts == nil || len(opts) < 1 {
		opts = []gqlgenHandler.Option{
			gqlgenHandler.CacheSize(s.config.GQLCacheSize),
			gqlgenHandler.ComplexityLimit(s.config.GQLComplexityLimit),
			// gqlgenHandler.EnablePersistedQueryCache(nil),
			gqlgenHandler.ErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
				// Refer to https://gqlgen.com/reference/errors/#the-error-presenter for custom error handling.
				return graphql.DefaultErrorPresenter(ctx, err)
			}),
			gqlgenHandler.RecoverFunc(func(ctx context.Context, err interface{}) error {
				// TODO: Implement error alert.
				return err.(error)
			}),
			gqlgenHandler.IntrospectionEnabled(s.config.GQLPlaygroundEnabled),
			// gqlgenHandler.RequestMiddleware(gqlapollotracing.RequestMiddleware()),
			// gqlgenHandler.Tracer(gqlapollotracing.NewTracer()),
			gqlgenHandler.UploadMaxMemory(s.config.GQLUploadMaxMemory),
			gqlgenHandler.UploadMaxSize(s.config.GQLUploadMaxSize),
			gqlgenHandler.WebsocketKeepAliveDuration(s.config.GQLWebsocketKeepAliveDuration),
		}
	}

	middleware = append(middleware, gqlHandler(schema, opts))
	s.router.Any(path, middleware...)

	if s.config.GQLPlaygroundEnabled && s.config.GQLPlaygroundPath != "" {
		s.router.GET(s.config.GQLPlaygroundPath, CSRFSkipCheck(), func(ctx *Context) {
			ctx.Data(http.StatusOK, "text/html", gqlPlaygroundTpl(path, ctx))
		})
	}
}

func gqlHandler(schema graphql.ExecutableSchema, opts []gqlgenHandler.Option) HandlerFunc {
	h := gqlgenHandler.GraphQL(schema, opts...)

	return func(ctx *Context) {
		h.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func gqlPlaygroundTpl(path string, ctx *Context) []byte {
	return []byte(`
<!DOCTYPE html>
<html>
<head>
	<meta charset=utf-8/>
	<meta name="viewport" content="user-scalable=no, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, minimal-ui">
	<title>GraphQL Playground</title>
	<link rel="stylesheet" href="//cdn.jsdelivr.net/npm/graphql-playground-react/build/static/css/index.css" />
	<link rel="shortcut icon" href="//cdn.jsdelivr.net/npm/graphql-playground-react/build/favicon.png" />
	<script src="//cdn.jsdelivr.net/npm/graphql-playground-react/build/static/js/middleware.js"></script>
</head>
<body>
	<div id="root">
	<style>
		body { background-color: rgb(23, 42, 58); font-family: Open Sans, sans-serif; height: 90vh; }
		#root { height: 100%; width: 100%; display: flex; align-items: center; justify-content: center; }
		.loading { font-size: 32px; font-weight: 200; color: rgba(255, 255, 255, .6); margin-left: 20px; }
		img { width: 78px; height: 78px; }
		.title { font-weight: 400; }
	</style>
	<img src='//cdn.jsdelivr.net/npm/graphql-playground-react/build/logo.png' alt=''>
	<div class="loading"> Loading
		<span class="title">GraphQL Playground</span>
	</div>
	</div>
	<script>
		function getCookie(name) {
			var v = document.cookie.match('(^|;) ?' + name + '=([^;]*)(;|$)');
			return v ? v[2] : null;
		}
		window.addEventListener('load', function (event) {
			GraphQLPlayground.init(document.getElementById('root'), {
				endpoint: '` + path + `',
				subscriptionEndpoint: '` + path + `',
				headers: {
					'X-CSRF-Token': unescape(getCookie("` + csrfTemplateFieldName(ctx) + `"))
				},
				settings: {
					'request.credentials': 'include',
					'schema.polling.interval': 5000
				}
			})
		})
	</script>
</body>
</html>
`)
}

// Router returns the server router instance.
func (s Server) Router() *Router {
	return s.router
}

// SetupRoutes configures routes for the server.
func (s Server) SetupRoutes(cb func(router *Router)) {
	cb(s.router)
}

// InitCSR initializes the client-side rendering/routing with index.html fallback.
func (s Server) InitCSR() {
	// Setup CSR hosting at "/".
	s.router.Use(s.serveCSR("/", s.assets))

	// Setup CSR hosting at "/tools".
	// s.Router.Use(s.serveCSR("/tools", tools.Assets))

	// Note: This might be not needed once we're sure everything is supposed to be on the PWA.
	s.router.NoRoute(CSRFSkipCheck(), func(ctx *Context) {
		ctx.HTML(http.StatusNotFound, "error/404", H{
			"title": "404 Page Not Found",
		})
	})
}

func (s Server) csrResource(path string) csrResource {
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

func (s *Server) serveCSR(prefix string, assets http.FileSystem) HandlerFunc {
	s.csrResources = append(s.csrResources, csrResource{
		assets:     assets,
		fileServer: http.StripPrefix(prefix, http.FileServer(assets)),
		prefix:     prefix,
	})

	return func(ctx *Context) {
		request := ctx.Request

		// Serve from the webpack-dev-server(debug) or assets(release) if the URL path isn't matching any of the SSR
		// paths.
		if !isSSRPath(s.Routes(), request.URL.Path) && !strings.HasPrefix(request.URL.Path, "/"+s.ssrPaths["root"]) {
			resource := s.csrResource(request.URL.Path)

			if appysupport.IsDebugBuild() {
				director := func(req *http.Request) {
					port, _ := strconv.Atoi(s.config.HTTPPort)
					req.URL.Scheme = "http"
					if s.config.HTTPSSLEnabled {
						port, _ = strconv.Atoi(s.config.HTTPSSLPort)
						req.URL.Scheme = "https"
					}

					hostname := s.config.HTTPHost + ":" + strconv.Itoa(port+1)
					req.URL.Host = hostname
					req.Host = hostname
				}
				proxy := &httputil.ReverseProxy{Director: director}
				proxy.ServeHTTP(ctx.Writer, ctx.Request)
				ctx.Abort()
				return
			}

			if resource.assets != nil {
				assetPath := strings.Replace(request.URL.Path, resource.prefix, "", 1)
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

// InitSSR initiates the SSR setup.
func (s Server) InitSSR() error {
	if err := s.initSSRLocale(); err != nil {
		return err
	}

	return s.initSSRView()
}

func (s *Server) initSSRLocale() error {
	s.i18nBundle = i18n.NewBundle(language.English)
	s.i18nBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	s.i18nBundle.RegisterUnmarshalFunc("yml", yaml.Unmarshal)
	s.i18nBundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	var (
		localeFiles []os.FileInfo
		data        []byte
		err         error
	)
	localeDir := s.ssrPaths["locale"]

	// Try getting all the locale files from `pkg/locales`, but fallback to `assets` http.FileSystem.
	if appysupport.IsDebugBuild() {
		localeFiles, err = ioutil.ReadDir(localeDir)

		if err != nil {
			return err
		}
	} else {
		localeDir = "/" + s.ssrPaths["root"] + "/" + localeDir
		file, err := s.assets.Open(localeDir)
		if err != nil {
			return err
		}

		localeFiles, _ = file.Readdir(-1)
	}

	for _, localeFile := range localeFiles {
		localeFn := localeFile.Name()

		if appysupport.IsDebugBuild() {
			data, _ = ioutil.ReadFile(localeDir + "/" + localeFn)
		} else {
			file, err := s.assets.Open(localeDir + "/" + localeFn)
			if err != nil {
				return err
			}

			data, _ = ioutil.ReadAll(file)
		}

		s.i18nBundle.MustParseMessageFileBytes(data, localeFn)
	}

	s.router.Use(I18n(s.i18nBundle))
	return nil
}

func (s *Server) initSSRView() error {
	var (
		fis []os.FileInfo
		err error
	)

	viewDir := s.ssrPaths["view"]

	// We will always read from local file system when it's debug build. Otherwise, read from the bind assets.
	if appysupport.IsDebugBuild() {
		if fis, err = ioutil.ReadDir(viewDir); err != nil {
			return err
		}
	} else {
		viewDir = "/" + s.ssrPaths["root"] + "/" + viewDir

		var file http.File
		if file, err = s.assets.Open(viewDir); err != nil {
			return err
		}

		fis, _ = file.Readdir(-1)
	}

	commonHTMLTpls := []string{}
	commonTextTpls := []string{}
	for _, fi := range fis {
		// We should only see directories in `pkg/views`.
		if fi.IsDir() == false {
			continue
		}

		if appysupport.ArrayContains(reservedViewDirs, fi.Name()) == true {
			htmlTpls, textTpls, err := getCommonTemplates(s.assets, viewDir+"/"+fi.Name())
			if err != nil {
				return err
			}

			commonHTMLTpls = append(commonHTMLTpls, htmlTpls...)
			commonTextTpls = append(commonTextTpls, textTpls...)
		}
	}

	for _, fi := range fis {
		if fi.IsDir() == false || appysupport.ArrayContains(reservedViewDirs, fi.Name()) == true {
			continue
		}

		var fileInfos []os.FileInfo
		targetDir := viewDir + "/" + fi.Name()
		if appysupport.IsDebugBuild() {
			fileInfos, _ = ioutil.ReadDir(targetDir)
		} else {
			var file http.File
			if file, err = s.assets.Open(targetDir); err != nil {
				return err
			}

			fileInfos, _ = file.Readdir(-1)
		}

		for _, fileInfo := range fileInfos {
			if fileInfo.IsDir() == true {
				continue
			}

			viewName := fi.Name() + "/" + fileInfo.Name()
			targetFn := targetDir + "/" + fileInfo.Name()
			data, err := getTemplateContent(s.assets, targetFn)
			if err != nil {
				return err
			}

			var commonTplsCopy []string
			if regexp.MustCompile(`\.html$`).Match([]byte(targetFn)) {
				commonTplsCopy = make([]string, len(commonHTMLTpls))
				copy(commonTplsCopy, commonHTMLTpls)
			} else if regexp.MustCompile(`\.txt$`).Match([]byte(targetFn)) {
				commonTplsCopy = make([]string, len(commonTextTpls))
				copy(commonTplsCopy, commonTextTpls)
			}

			viewContent := append(commonTplsCopy, data)
			s.htmlRenderer.AddFromStringsFuncs(viewName, s.viewHelper, viewContent...)
		}
	}

	return nil
}

func getCommonTemplates(assets http.FileSystem, path string) ([]string, []string, error) {
	var (
		fis []os.FileInfo
		err error
	)

	htmlTpls := []string{}
	textTpls := []string{}
	if appysupport.IsDebugBuild() {
		fis, _ = ioutil.ReadDir(path)
	} else {
		var file http.File
		path = "/" + path

		if file, err = assets.Open(path); err != nil {
			return nil, nil, err
		}

		fis, _ = file.Readdir(-1)
	}

	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		data, err := getTemplateContent(assets, path+"/"+fi.Name())
		if err != nil {
			return nil, nil, err
		}

		if regexp.MustCompile(`\.html$`).Match([]byte(fi.Name())) {
			htmlTpls = append(htmlTpls, data)
		} else if regexp.MustCompile(`\.txt$`).Match([]byte(fi.Name())) {
			textTpls = append(textTpls, data)
		}
	}

	return htmlTpls, textTpls, nil
}

func getTemplateContent(assets http.FileSystem, path string) (string, error) {
	var data []byte
	if appysupport.IsDebugBuild() {
		data, _ := ioutil.ReadFile(path)
		return string(data), nil
	}

	file, err := assets.Open(path)
	if err != nil {
		return "", err
	}

	data, _ = ioutil.ReadAll(file)
	return string(data), nil
}

func isSSRPath(routes []RouteInfo, path string) bool {
	for _, route := range routes {
		if route.Path == path {
			return true
		}
	}

	return false
}
