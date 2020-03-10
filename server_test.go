package appy_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/appist/appy"
	"github.com/gorilla/websocket"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type ServerSuite struct {
	appy.TestSuite
	asset   *appy.Asset
	config  *appy.Config
	logger  *appy.Logger
	buffer  *bytes.Buffer
	writer  *bufio.Writer
	support appy.Supporter
}

func (s *ServerSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.support = &appy.Support{}
	s.asset = appy.NewAsset(nil, map[string]string{
		"docker": "testdata/server/.docker",
		"config": "testdata/server/configs",
		"locale": "testdata/server/pkg/locales",
		"view":   "testdata/server/pkg/views",
		"web":    "testdata/server/web",
	}, "")
	s.logger, s.buffer, s.writer = appy.NewFakeLogger()
	s.config = appy.NewConfig(s.asset, s.logger, s.support)
}

func (s *ServerSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *ServerSuite) TestNewServerWithSSLEnabled() {
	s.config.HTTPSSLEnabled = true
	server := appy.NewServer(s.asset, s.config, s.logger, s.support)
	s.NotNil(server.Config())
	s.NotNil(server.HTTP())
	s.NotNil(server.HTTPS())
	s.NotNil(server.HTMLRenderer())
	s.Equal(0, len(server.Middleware()))
	s.Equal("localhost:3000", server.HTTP().Addr)
	s.Equal("localhost:3443", server.HTTPS().Addr)
}

func (s *ServerSuite) TestNewServerWithoutSSLEnabled() {
	server := appy.NewServer(s.asset, s.config, s.logger, s.support)
	s.NotNil(server.Config())
	s.NotNil(server.HTTP())
	s.NotNil(server.HTTPS())
	s.NotNil(server.HTMLRenderer())
	s.Equal(0, len(server.Middleware()))
	s.Equal("localhost:3000", server.HTTP().Addr)
	s.Equal("localhost:3443", server.HTTPS().Addr)
}

func (s *ServerSuite) TestIsSSLCertsExisted() {
	server := appy.NewServer(s.asset, s.config, s.logger, s.support)
	s.Equal(false, server.IsSSLCertExisted())

	s.config.HTTPSSLCertPath = "testdata/server/ssl"
	server = appy.NewServer(s.asset, s.config, s.logger, s.support)
	s.Equal(true, server.IsSSLCertExisted())
}

func (s *ServerSuite) TestInfo() {
	server := appy.NewServer(s.asset, s.config, s.logger, s.support)
	output := server.Info()
	s.Contains(output, fmt.Sprintf("* appy 0.1.0 (%s), build: debug, environment: development, config: testdata/server/configs/.env.development", runtime.Version()))
	s.Contains(output, "* Listening on http://localhost:3000")

	s.config.HTTPSSLEnabled = true
	server = appy.NewServer(s.asset, s.config, s.logger, s.support)
	output = server.Info()
	s.Contains(output, fmt.Sprintf("* appy 0.1.0 (%s), build: debug, environment: development, config: testdata/server/configs/.env.development", runtime.Version()))
	s.Contains(output, "* Listening on http://localhost:3000, https://localhost:3443")

	s.config.HTTPHost = "0.0.0.0"
	server = appy.NewServer(s.asset, s.config, s.logger, s.support)
	output = server.Info()
	s.Contains(output, fmt.Sprintf("* appy 0.1.0 (%s), build: debug, environment: development, config: testdata/server/configs/.env.development", runtime.Version()))
	s.Contains(output, "* Listening on http://0.0.0.0:3000, https://0.0.0.0:3443")
}

func (s *ServerSuite) TestRouting() {
	server := appy.NewServer(s.asset, s.config, s.logger, s.support)
	server.ServeNoRoute()
	s.Equal(server.BasePath(), "/")

	w := server.TestHTTPRequest("GET", "/foobar", nil, nil)
	defer w.Close()

	s.Equal(http.StatusNotFound, w.Code)
	s.Contains(w.Body.String(), "<title>404 Page Not Found</title>")

	server.Any("/foo", func(c *appy.Context) { c.String(http.StatusOK, "bar") })
	methods := []string{"CONNECT", "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "PUT", "POST", "TRACE"}

	for _, method := range methods {
		w := server.TestHTTPRequest(method, "/foo", nil, nil)
		defer w.Close()

		s.Equal(http.StatusOK, w.Code)
		s.Equal("bar", w.Body.String())
	}

	server.Handle("CONNECT", "/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	server.DELETE("/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	server.GET("/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	server.HEAD("/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	server.OPTIONS("/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	server.PATCH("/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	server.PUT("/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	server.POST("/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	methods = []string{"CONNECT", "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "PUT", "POST"}

	for _, method := range methods {
		w := server.TestHTTPRequest(method, "/bar", nil, nil)
		defer w.Close()

		s.Equal(http.StatusOK, w.Code)
		s.Equal("foo", w.Body.String())
	}

	count := 1
	v1 := server.Group("/v1")
	v1.Use(func(c *appy.Context) {
		count = 10
		c.Next()
	})
	v1.Any("/foo", func(c *appy.Context) { c.String(http.StatusOK, "bar") })
	methods = []string{"CONNECT", "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "PUT", "POST", "TRACE"}

	for _, method := range methods {
		w := server.TestHTTPRequest(method, "/v1/foo", nil, nil)
		defer w.Close()

		s.Equal(http.StatusOK, w.Code)
		s.Equal("bar", w.Body.String())
	}
	s.Equal(10, count)

	v1.Handle("CONNECT", "/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	v1.DELETE("/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	v1.GET("/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	v1.HEAD("/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	v1.OPTIONS("/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	v1.PATCH("/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	v1.PUT("/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	v1.POST("/bar", func(c *appy.Context) { c.String(http.StatusOK, "foo") })
	methods = []string{"CONNECT", "DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "PUT", "POST"}

	for _, method := range methods {
		w := server.TestHTTPRequest(method, "/v1/bar", nil, nil)
		defer w.Close()

		s.Equal(http.StatusOK, w.Code)
		s.Equal("foo", w.Body.String())
	}

	routes := server.Routes()
	s.Equal(35, len(routes))

	route := routes[len(routes)-2]
	recorder := httptest.NewRecorder()
	c, _ := appy.NewTestContext(recorder)
	s.Equal("/v1/foo", route.Path)

	route.HandlerFunc(c)
	s.Equal(http.StatusOK, recorder.Code)
	s.Equal("bar", recorder.Body.String())
}

func (s *ServerSuite) TestCSRWithDebugBuild() {
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

	server := appy.NewServer(s.asset, s.config, s.logger, s.support)
	server.ServeSPA("/", nil)
	w := server.TestHTTPRequest("GET", "/", nil, nil)
	defer w.Close()

	// Since reverse proxy is working in test and the webpack-dev-server not running, it should throw 502.
	s.Equal(http.StatusBadGateway, w.Code)

	s.config.HTTPSSLEnabled = true
	server = appy.NewServer(s.asset, s.config, s.logger, s.support)
	server.ServeSPA("/", nil)
	w = server.TestHTTPRequest("GET", "/", nil, nil)

	// Since reverse proxy is working in test and the webpack-dev-server not running, it should throw 502.
	s.Equal(http.StatusBadGateway, w.Code)
}

func (s *ServerSuite) TestCSRWithReleaseBuild() {
	appy.Build = appy.ReleaseBuild
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "58f364f29b568807ab9cffa22c99b538")
	os.Setenv("HTTP_CSRF_SECRET", "b13f61fe0411522f25f3d60a7588ffeecbcde8146f193fa3406eee81ad67b5ec5ad5619a9a4aa5975a20ad1f911489ec74d096e584251cb728d4b0b5")
	os.Setenv("HTTP_SESSION_SECRETS", "b13f61fe0411522f25f3d60a7588ffeecbcde8146f193fa3406eee81ad67b5ec5ad5619a9a4aa5975a20ad1f911489ec74d096e584251cb728d4b0b5")
	defer func() {
		appy.Build = appy.DebugBuild
		os.Unsetenv("APPY_ENV")
		os.Unsetenv("APPY_MASTER_KEY")
		os.Unsetenv("HTTP_CSRF_SECRET")
		os.Unsetenv("HTTP_SESSION_SECRETS")
	}()

	server := appy.NewServer(s.asset, s.config, s.logger, s.support)
	server.ServeSPA("/", http.Dir("testdata/server"))
	w := server.TestHTTPRequest("GET", "/", nil, nil)
	defer w.Close()

	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), `<div id="app">we build apps</div>`)

	server = appy.NewServer(s.asset, s.config, s.logger, s.support)
	server.ServeSPA("/", nil)
	w = server.TestHTTPRequest("GET", "/", nil, nil)

	s.Equal(http.StatusNotFound, w.Code)

	server = appy.NewServer(s.asset, s.config, s.logger, s.support)
	server.GET("/ssr", func(c *appy.Context) {
		c.String(http.StatusOK, "foobar")
	})
	server.ServeSPA("/", http.Dir("testdata/server"))
	server.ServeSPA("/ssr", http.Dir("testdata/server"))
	server.ServeSPA("/tools", http.Dir("testdata/server/tools"))
	server.ServeNoRoute()

	w = server.TestHTTPRequest("GET", "/ssr/foo", nil, nil)
	s.Equal(http.StatusNotFound, w.Code)
	s.Contains(w.Body.String(), "<title>404 Page Not Found</title>")

	w = server.TestHTTPRequest("GET", "/foo", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), `<div id="app">we build apps</div>`)

	w = server.TestHTTPRequest("GET", "/tools", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), `<div id="app">we build another SPA</div>`)

	w = server.TestHTTPRequest("GET", "/ssr", nil, nil)
	s.Equal(http.StatusOK, w.Code)
	s.Contains(w.Body.String(), "foobar")

	routes := []string{"/" + s.asset.Layout()["config"], "/" + s.asset.Layout()["locale"], "/" + s.asset.Layout()["view"]}
	for _, route := range routes {
		w = server.TestHTTPRequest("GET", route, nil, nil)
		s.Equal(http.StatusNotFound, w.Code)
		s.Contains(w.Body.String(), "<title>404 Page Not Found</title>")
	}

	server = appy.NewServer(s.asset, s.config, s.logger, s.support)
	server.ServeSPA("/", nil)
	server.ServeNoRoute()

	w = server.TestHTTPRequest("GET", "/", nil, nil)
	s.Equal(http.StatusNotFound, w.Code)
	s.Contains(w.Body.String(), "<title>404 Page Not Found</title>")
}

type fakeGQLExt struct{}

var _ interface {
	graphql.OperationParameterMutator
	graphql.HandlerExtension
} = fakeGQLExt{}

func (c fakeGQLExt) ExtensionName() string {
	return "fakeGQLExt"
}

func (c fakeGQLExt) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (c fakeGQLExt) MutateOperationParameters(ctx context.Context, rawParams *graphql.RawParams) *gqlerror.Error {
	return nil
}

func (s *ServerSuite) TestSetupGraphQL() {
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

	s.config.GQLPlaygroundEnabled = true
	s.config.GQLPlaygroundPath = "/graphiql"
	graphqlPath := "/graphql"
	server := appy.NewServer(s.asset, s.config, s.logger, s.support)
	server.Use(appy.CSRF(s.config, s.logger, s.support))
	server.SetupGraphQL(graphqlPath, nil, []graphql.HandlerExtension{fakeGQLExt{}})

	w := server.TestHTTPRequest("GET", "/graphiql", nil, nil)
	defer w.Close()

	s.Equal(200, w.Code)
	s.Contains(w.Body.String(), "<title>GraphQL Playground</title>")

	w = server.TestHTTPRequest("POST", graphqlPath, nil, nil)
	s.Equal(403, w.Code)

	w = server.TestHTTPRequest("POST", graphqlPath, appy.H{
		"content-type": "application/json",
		"x-api-only":   "1",
	}, nil)
	s.Equal(422, w.Code)

	ts := httptest.NewServer(server.Router())
	ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(ts.URL, "http")+graphqlPath, nil)
	defer ws.Close()
	s.Nil(err)

	ws, _, err = websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(ts.URL, "http"), nil)
	s.NotNil(err)
}

func TestServerSuite(t *testing.T) {
	appy.RunTestSuite(t, new(ServerSuite))
}
