package http

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"

	appysupport "github.com/appist/appy/internal/support"
	"github.com/appist/appy/internal/test"
)

type (
	ServerSuite struct {
		test.Suite
		assets      http.FileSystem
		logger      *appysupport.Logger
		buffer      *bytes.Buffer
		writer      *bufio.Writer
		viewHelper  template.FuncMap
		oldSSRPaths map[string]string
		recorder    *TestResponseRecorder
	}

	TestResponseRecorder struct {
		*httptest.ResponseRecorder
		closeChannel chan bool
	}
)

func (r *TestResponseRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func (r *TestResponseRecorder) closeClient() {
	r.closeChannel <- true
}

func CreateTestResponseRecorder() *TestResponseRecorder {
	return &TestResponseRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

func (s *ServerSuite) SetupTest() {
	s.assets = http.Dir("testdata")
	s.logger, s.buffer, s.writer = appysupport.NewFakeLogger()
	s.viewHelper = template.FuncMap{
		"testViewHelper": func() string {
			return "i am view helper"
		},
	}
	s.recorder = CreateTestResponseRecorder()
	s.oldSSRPaths = appysupport.SSRPaths
	appysupport.SSRPaths = map[string]string{
		"root":   "testdata/.ssr",
		"config": "testdata/pkg/config",
		"locale": "testdata/pkg/locales",
		"view":   "testdata/pkg/views",
	}
}

func (s *ServerSuite) TearDownTest() {
	appysupport.SSRPaths = s.oldSSRPaths
}

func (s *ServerSuite) TestNewServerWithoutSSLEnabled() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	config := appysupport.NewConfig(s.logger, nil)
	s.Nil(config.Errors())

	server := NewServer(config, s.logger, s.assets, nil)
	s.NotNil(server.assets)
	s.NotNil(server.config)
	s.NotNil(server.http)
	s.NotNil(server.htmlRenderer)
	s.NotNil(server.router)
	s.Equal("localhost:3000", server.http.Addr)
	s.Equal("localhost:3443", server.https.Addr)
}

func (s *ServerSuite) TestNewServerWithSSLEnabled() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	config := appysupport.NewConfig(s.logger, nil)
	config.HTTPSSLEnabled = true
	server := NewServer(config, s.logger, s.assets, nil)
	s.NotNil(server.assets)
	s.NotNil(server.config)
	s.NotNil(server.http)
	s.NotNil(server.htmlRenderer)
	s.NotNil(server.router)
	s.Equal("localhost:3000", server.http.Addr)
	s.Equal("localhost:3443", server.https.Addr)
}

func (s *ServerSuite) TestInitSSRWithDebugBuild() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	config := appysupport.NewConfig(s.logger, nil)
	server := NewServer(config, s.logger, s.assets, s.viewHelper)
	s.NoError(server.InitSSR())

	server.router.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "welcome/index", nil)
	})
	request, _ := http.NewRequest("GET", "/", nil)
	server.router.ServeHTTP(s.recorder, request)
	s.Equal(500, s.recorder.Code)

	server = NewServer(config, s.logger, s.assets, s.viewHelper)
	s.NoError(server.InitSSR())
	server.router.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "welcome/index.html", H{"message": "i am testing"})
	})
	s.recorder = CreateTestResponseRecorder()
	request, _ = http.NewRequest("GET", "/", nil)
	server.router.ServeHTTP(s.recorder, request)

	s.Equal(200, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), "i am testing")
	s.Contains(s.recorder.Body.String(), "i am view helper")
}

func (s *ServerSuite) TestInitSSRWithReleaseBuild() {
	appysupport.Build = appysupport.ReleaseBuild
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		appysupport.Build = appysupport.DebugBuild
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	appysupport.SSRPaths = map[string]string{
		"root":   ".ssr",
		"config": "config",
		"locale": "locales",
		"view":   "views",
	}
	config := appysupport.NewConfig(s.logger, nil)
	server := NewServer(config, s.logger, s.assets, s.viewHelper)
	s.NoError(server.InitSSR())

	server.router.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "welcome/index", nil)
	})
	request, _ := http.NewRequest("GET", "/", nil)
	server.router.ServeHTTP(s.recorder, request)
	s.Equal(500, s.recorder.Code)

	server = NewServer(config, s.logger, s.assets, s.viewHelper)
	s.NoError(server.InitSSR())
	server.router.GET("/", func(c *Context) {
		c.HTML(http.StatusOK, "welcome/index.html", H{"message": "i am testing"})
	})
	s.recorder = CreateTestResponseRecorder()
	request, _ = http.NewRequest("GET", "/", nil)
	server.router.ServeHTTP(s.recorder, request)

	s.Equal(200, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), "i am testing")
	s.Contains(s.recorder.Body.String(), "i am view helper")
}

func (s *ServerSuite) TestIsSSLCertsExisted() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	config := appysupport.NewConfig(s.logger, nil)
	server := NewServer(config, s.logger, s.assets, s.viewHelper)
	s.Equal(false, server.IsSSLCertExisted())

	config.HTTPSSLCertPath = "testdata/ssl"
	server = NewServer(config, s.logger, s.assets, nil)
	s.Equal(true, server.IsSSLCertExisted())
}

func (s *ServerSuite) TestCSRAssetsNotConfiguredWithoutSSLEnabled() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "b13f61fe0411522f25f3d60a7588ffeecbcde8146f193fa3406eee81ad67b5ec5ad5619a9a4aa5975a20ad1f911489ec74d096e584251cb728d4b0b5")
	os.Setenv("HTTP_SESSION_SECRETS", "b13f61fe0411522f25f3d60a7588ffeecbcde8146f193fa3406eee81ad67b5ec5ad5619a9a4aa5975a20ad1f911489ec74d096e584251cb728d4b0b5")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	request, _ := http.NewRequest("GET", "/", nil)
	config := appysupport.NewConfig(s.logger, nil)
	server := NewServer(config, s.logger, nil, nil)
	server.InitCSR()
	server.router.ServeHTTP(s.recorder, request)

	// Since reverse proxy is working in test and the webpack-dev-server not running, it should throw 502.
	s.Equal(502, s.recorder.Code)
}

func (s *ServerSuite) TestCSRAssetsNotConfiguredWithSSLEnabled() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "b13f61fe0411522f25f3d60a7588ffeecbcde8146f193fa3406eee81ad67b5ec5ad5619a9a4aa5975a20ad1f911489ec74d096e584251cb728d4b0b5")
	os.Setenv("HTTP_SESSION_SECRETS", "b13f61fe0411522f25f3d60a7588ffeecbcde8146f193fa3406eee81ad67b5ec5ad5619a9a4aa5975a20ad1f911489ec74d096e584251cb728d4b0b5")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	request, _ := http.NewRequest("GET", "/", nil)
	config := appysupport.NewConfig(s.logger, nil)
	config.HTTPSSLEnabled = true
	server := NewServer(config, s.logger, nil, nil)
	server.InitCSR()
	server.router.ServeHTTP(s.recorder, request)

	// Since reverse proxy is working in test and the webpack-dev-server not running, it should throw 502.
	s.Equal(502, s.recorder.Code)
}

func (s *ServerSuite) TestNonExistingPathWithCSRAssetsPresent() {
	appysupport.Build = appysupport.ReleaseBuild
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "b13f61fe0411522f25f3d60a7588ffeecbcde8146f193fa3406eee81ad67b5ec5ad5619a9a4aa5975a20ad1f911489ec74d096e584251cb728d4b0b5")
	os.Setenv("HTTP_SESSION_SECRETS", "b13f61fe0411522f25f3d60a7588ffeecbcde8146f193fa3406eee81ad67b5ec5ad5619a9a4aa5975a20ad1f911489ec74d096e584251cb728d4b0b5")
	defer func() {
		appysupport.Build = appysupport.DebugBuild
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	request, _ := http.NewRequest("GET", "/dummy", nil)
	config := appysupport.NewConfig(s.logger, nil)
	server := NewServer(config, s.logger, http.Dir("testdata/csr"), nil)
	server.InitCSR()
	server.router.ServeHTTP(s.recorder, request)

	s.Equal(200, s.recorder.Code)
	s.Contains(s.recorder.Body.String(), "<div id=\"app\">we build apps</div>")
}

func (s *ServerSuite) TestServerPrintInfoWithDebugBuild() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "b13f61fe0411522f25f3d60a7588ffeecbcde8146f193fa3406eee81ad67b5ec5ad5619a9a4aa5975a20ad1f911489ec74d096e584251cb728d4b0b5")
	os.Setenv("HTTP_SESSION_SECRETS", "b13f61fe0411522f25f3d60a7588ffeecbcde8146f193fa3406eee81ad67b5ec5ad5619a9a4aa5975a20ad1f911489ec74d096e584251cb728d4b0b5")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	config := appysupport.NewConfig(s.logger, nil)
	server := NewServer(config, s.logger, nil, nil)
	output := server.Info()
	s.Contains(output, fmt.Sprintf("* appy 0.1.0 (%s), build: debug, environment: development, config: testdata/pkg/config/.env.development", runtime.Version()))
	s.Contains(output, "* Listening on http://localhost:3000")

	os.Setenv("HTTP_SSL_ENABLED", "true")
	defer func() {
		os.Unsetenv("HTTP_SSL_ENABLED")
	}()
	config = appysupport.NewConfig(s.logger, http.Dir("testdata"))
	server = NewServer(config, s.logger, http.Dir("./testdata"), nil)
	output = server.Info()
	s.Contains(output, fmt.Sprintf("* appy 0.1.0 (%s), build: debug, environment: development, config: testdata/pkg/config/.env.development", runtime.Version()))
	s.Contains(output, "* Listening on http://localhost:3000, https://localhost:3443")

	os.Setenv("HTTP_HOST", "0.0.0.0")
	defer func() {
		os.Unsetenv("HTTP_HOST")
	}()
	config = appysupport.NewConfig(s.logger, http.Dir("testdata"))
	server = NewServer(config, s.logger, http.Dir("testdata"), nil)
	output = server.Info()
	s.Contains(output, fmt.Sprintf("* appy 0.1.0 (%s), build: debug, environment: development, config: testdata/pkg/config/.env.development", runtime.Version()))
	s.Contains(output, "* Listening on http://0.0.0.0:3000, https://0.0.0.0:3443")
}

func (s *ServerSuite) TestSetRoutes() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	defer func() {
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	config := appysupport.NewConfig(s.logger, nil)
	server := NewServer(config, s.logger, nil, nil)
	server.SetupRoutes(func(r *Router) {
		r.GET("/dummy", func(ctx *Context) {})
	})

	s.Equal("/dummy", server.Routes()[0].Path)
}

func TestServerSuite(t *testing.T) {
	test.RunSuite(t, new(ServerSuite))
}
