package core

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"

	"github.com/appist/appy/test"
)

type ServerSuite struct {
	test.Suite
	assets     http.FileSystem
	config     AppConfig
	logger     *AppLogger
	viewHelper template.FuncMap
}

func (s *ServerSuite) SetupTest() {
	Build = "debug"
	s.assets = http.Dir("./testdata")
	s.config, _ = newConfig(s.assets)
	s.config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	s.logger, _ = newLogger(newLoggerConfig())
	s.viewHelper = template.FuncMap{
		"testViewHelper": func() string {
			return "i am view helper"
		},
	}
}

func (s *ServerSuite) TearDownTest() {
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_HOST")
	os.Unsetenv("HTTP_SESSION_SECRETS")
	os.Unsetenv("HTTP_SSL_ENABLED")
}

func (s *ServerSuite) TestNewServerWithoutSSLEnabled() {
	server := newServer(s.assets, s.config, s.logger, nil)
	s.NotNil(server.assets)
	s.NotNil(server.Config)
	s.NotNil(server.HTTP)
	s.NotNil(server.htmlRenderer)
	s.NotNil(server.Router)
	s.Equal("localhost:3000", server.HTTP.Addr)
}

func (s *ServerSuite) TestNewServerWithSSLEnabled() {
	s.config.HTTPSSLEnabled = true
	server := newServer(s.assets, s.config, s.logger, nil)
	s.NotNil(server.assets)
	s.NotNil(server.Config)
	s.NotNil(server.HTTP)
	s.NotNil(server.htmlRenderer)
	s.NotNil(server.Router)
	s.Equal("localhost:3443", server.HTTP.Addr)
}

func (s *ServerSuite) TestDefaultWelcomePageWithoutCustomHomePath() {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/", nil)
	server := newServer(s.assets, s.config, s.logger, nil)
	server.AddDefaultWelcomePage()
	server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Equal("text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Contains(recorder.Body.String(), "<p class=\"lead\">An opinionated productive web framework that helps scaling business easier.</p>")
}

func (s *ServerSuite) TestDefaultWelcomePageWithCustomHomePath() {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/", nil)
	server := newServer(s.assets, s.config, s.logger, nil)
	server.Router.GET("/", func(c *Context) {
		c.JSON(200, H{"a": 1})
	})
	server.AddDefaultWelcomePage()
	server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Equal("application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Equal("{\"a\":1}\n", recorder.Body.String())
}

func (s *ServerSuite) TestDefaultWelcomePageWithCSRHomePath() {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/", nil)
	server := newServer(http.Dir("./testdata/.ssr"), s.config, s.logger, nil)
	server.AddDefaultWelcomePage()
	server.InitCSR()
	server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Equal("text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Contains(recorder.Body.String(), "we build apps")
}

func (s *ServerSuite) TestIsSSLCertsExist() {
	server := newServer(s.assets, s.config, s.logger, nil)
	s.Equal(false, server.IsSSLCertsExist())

	s.config.HTTPSSLCertPath = "./testdata/ssl"
	server = newServer(s.assets, s.config, s.logger, nil)
	s.Equal(true, server.IsSSLCertsExist())
}

func (s *ServerSuite) TestCSRAssetsNotConfigured() {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/", nil)
	server := newServer(nil, s.config, s.logger, nil)
	server.InitCSR()
	server.Router.ServeHTTP(recorder, request)

	s.Equal(404, recorder.Code)
}

func (s *ServerSuite) TestNonExistingPathWithCSRAssetsNil() {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/dummy", nil)
	server := newServer(nil, s.config, s.logger, nil)
	server.InitCSR()
	server.Router.ServeHTTP(recorder, request)

	s.Equal(404, recorder.Code)
}

func (s *ServerSuite) TestNonExistingPathWithCSRAssetsPresent() {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/dummy", nil)
	server := newServer(http.Dir("./testdata/csr"), s.config, s.logger, nil)
	server.InitCSR()
	server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Contains(recorder.Body.String(), "<div id=\"app\">we build apps</div>")
}

func (s *ServerSuite) TestStaticAssets301Redirect() {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/index.html", nil)
	server := newServer(http.Dir("./testdata/csr"), s.config, s.logger, nil)
	server.InitCSR()
	server.Router.ServeHTTP(recorder, request)

	s.Equal(301, recorder.Code)
}

func (s *ServerSuite) TestSSRWithCSRInitialized() {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/welcome", nil)
	server := newServer(http.Dir("./testdata/csr"), s.config, s.logger, nil)
	server.Router.GET("/welcome", func(c *Context) {
		c.String(200, "%s", "test")
	})
	server.InitCSR()
	server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Equal("test", recorder.Body.String())
}

func (s *ServerSuite) TestSSRWithAdditionalCSRInitialized() {
	server := newServer(http.Dir("./testdata/csr"), s.config, s.logger, nil)
	server.serveCSR("/tools", http.Dir("./testdata/tools"))
	server.InitCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/tools", nil)
	server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Contains(recorder.Body.String(), "<div id=\"app\">we build tools</div>")

	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/tools/dummy", nil)
	server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Contains(recorder.Body.String(), "<div id=\"app\">we build tools</div>")
}

func (s *ServerSuite) TestSSRWorksWithCSR() {
	server := newServer(http.Dir("./testdata/csr"), s.config, s.logger, nil)
	server.serveCSR("/tools", http.Dir("./testdata/tools"))
	server.InitCSR()
	server.Router.GET("/welcome", func(c *Context) {
		c.String(200, "%s", "test")
	})

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/welcome", nil)
	server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Equal("test", recorder.Body.String())
}

func (s *ServerSuite) TestServerPrintInfoWithDebugBuild() {
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	server := newServer(nil, s.config, s.logger, nil)
	output := CaptureOutput(func() {
		server.PrintInfo()
	})

	s.Contains(output, fmt.Sprintf("* Version 0.1.0 (%s), build: debug, environment: development, config: none", runtime.Version()))
	s.Contains(output, "* Listening on http://localhost:3000")

	os.Setenv("HTTP_SSL_ENABLED", "true")
	s.config, _ = newConfig(http.Dir("./testdata"))
	server = newServer(http.Dir("./testdata"), s.config, s.logger, nil)
	output = CaptureOutput(func() {
		server.PrintInfo()
	})

	s.Contains(output, fmt.Sprintf("* Version 0.1.0 (%s), build: debug, environment: development, config: none", runtime.Version()))
	s.Contains(output, "* Listening on https://localhost:3443")

	os.Setenv("HTTP_HOST", "0.0.0.0")
	s.config, _ = newConfig(http.Dir("./testdata"))
	server = newServer(http.Dir("./testdata"), s.config, s.logger, nil)
	output = CaptureOutput(func() {
		server.PrintInfo()
	})

	s.Contains(output, fmt.Sprintf("* Version 0.1.0 (%s), build: debug, environment: development, config: none", runtime.Version()))
	s.Contains(output, "* Listening on https://0.0.0.0:3443")
}

func (s *ServerSuite) TestServerPrintInfoWithReleaseBuild() {
	Build = "release"
	s.config, _ = newConfig(http.Dir("./testdata/.ssr"))
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	server := newServer(nil, s.config, s.logger, nil)
	output := CaptureOutput(func() {
		server.PrintInfo()
	})

	// Release build should use AppLogger instead of stdout
	s.Equal(output, "")
	s.Equal(output, "")

	os.Setenv("HTTP_SSL_ENABLED", "true")
	server = newServer(nil, s.config, s.logger, nil)
	output = CaptureOutput(func() {
		server.PrintInfo()
	})

	// Release build should use AppLogger instead of stdout
	s.Equal(output, "")
	s.Equal(output, "")
}

func (s *ServerSuite) TestInitSSRWithDebugBuildWithCorrectPath() {
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	s.config, _ = newConfig(http.Dir("./testdata/.ssr"))

	recorder := httptest.NewRecorder()
	server := newServer(nil, s.config, s.logger, s.viewHelper)
	oldSSRLocale := server.SSRPaths["locale"]
	oldSSRView := server.SSRPaths["view"]
	server.SSRPaths["locale"] = "./testdata/.ssr/app/locales"
	server.SSRPaths["view"] = "./testdata/.ssr/app/views"
	s.NoError(server.InitSSR())
	server.Router.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "welcome/index", nil)
	})
	request, _ := http.NewRequest("GET", "/", nil)
	server.Router.ServeHTTP(recorder, request)

	s.Equal(500, recorder.Code)

	server = newServer(nil, s.config, s.logger, s.viewHelper)
	s.NoError(server.InitSSR())
	server.Router.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "welcome/index.html", H{"message": "i am testing"})
	})
	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/", nil)
	server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Contains(recorder.Body.String(), "i am testing")
	s.Contains(recorder.Body.String(), "i am view helper")

	server.SSRPaths["locale"] = oldSSRLocale
	server.SSRPaths["view"] = oldSSRView
}

func (s *ServerSuite) TestInitSSRWithDebugBuildWithIncorrectPath() {
	server := newServer(nil, AppConfig{}, s.logger, s.viewHelper)
	oldSSRLocale := server.SSRPaths["locale"]
	oldSSRView := server.SSRPaths["view"]
	server.SSRPaths["locale"] = "./testdata/.ssr_only_locales/app/locales"
	server.SSRPaths["view"] = "./testdata/.ssr_only_locales/app/views"
	s.EqualError(server.InitSSR(), fmt.Sprintf("open %s: no such file or directory", server.SSRPaths["view"]))

	server = newServer(nil, AppConfig{}, s.logger, s.viewHelper)
	server.SSRPaths["locale"] = "./testdata/.ssr_only_views/app/locales"
	server.SSRPaths["view"] = "./testdata/.ssr_only_views/app/views"
	s.EqualError(server.InitSSR(), fmt.Sprintf("open %s: no such file or directory", server.SSRPaths["locale"]))

	server.SSRPaths["locale"] = oldSSRLocale
	server.SSRPaths["view"] = oldSSRView
}

func (s *ServerSuite) TestInitSSRWithReleaseBuildWithCorrectPath() {
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	Build = "release"
	s.config, _ = newConfig(http.Dir("./testdata"))

	server := newServer(http.Dir("./testdata"), s.config, s.logger, s.viewHelper)
	s.NoError(server.InitSSR())
	server.Router.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "welcome/index", nil)
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/", nil)
	server.Router.ServeHTTP(recorder, request)

	s.Equal(500, recorder.Code)

	server = newServer(http.Dir("./testdata"), s.config, s.logger, s.viewHelper)
	s.NoError(server.InitSSR())
	server.Router.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "welcome/index.html", H{"message": "i am testing"})
	})
	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/", nil)
	server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Contains(recorder.Body.String(), "i am testing")
	s.Contains(recorder.Body.String(), "i am view helper")
}

func (s *ServerSuite) TestInitSSRWithReleaseBuildWithIncorrectPath() {
	Build = "release"

	server := newServer(http.Dir("./testdata/.ssr_only_locales"), AppConfig{}, s.logger, s.viewHelper)
	oldSSRLocale := server.SSRPaths["locale"]
	oldSSRView := server.SSRPaths["view"]
	server.SSRPaths["locale"] = "../app/locales"
	server.SSRPaths["view"] = "../app/views"
	s.EqualError(server.InitSSR(), "open testdata/.ssr_only_locales/app/views: no such file or directory")

	server = newServer(http.Dir("./testdata/.ssr_only_views"), AppConfig{}, s.logger, s.viewHelper)
	s.EqualError(server.InitSSR(), "open testdata/.ssr_only_views/app/locales: no such file or directory")

	server.SSRPaths["locale"] = oldSSRLocale
	server.SSRPaths["view"] = oldSSRView
}

func TestServer(t *testing.T) {
	test.Run(t, new(ServerSuite))
}
