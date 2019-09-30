package http

import (
	"appist/appy/support"
	"appist/appy/test"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDynamicRouting(t *testing.T) {
	assert := test.NewAssert(t)
	s := NewServer(support.Config)
	handler := func() gin.HandlerFunc { return func(c *gin.Context) {} }
	s.GET("/users/:id", handler())
	s.HEAD("/users/:id", handler())
	s.OPTIONS("/users/:id", handler())
	s.PATCH("/users/:id", handler())
	s.PUT("/users/:id", handler())
	s.POST("/login", handler())
	s.DELETE("/logout", handler())
	s.Any("/posts", handler())
	s.Handle("POST", "/register", handler())
	apiV1 := s.Group("/api/v1")
	{
		apiV1.GET("/users/:id", handler())
	}

	assert.Equal("/", s.BasePath())

	middlewares := s.Middlewares()
	assert.GreaterOrEqual(len(middlewares), 1)

	s.Use(func(c *gin.Context) {})
	assert.Greater(len(s.Middlewares()), len(middlewares))

	rs := map[string]bool{}
	for _, route := range s.Routes() {
		key := fmt.Sprintf("%s %s", route.Method, route.Path)
		rs[key] = true
	}

	tests := []string{
		"GET /users/:id",
		"OPTIONS /users/:id",
		"HEAD /users/:id",
		"PATCH /users/:id",
		"PUT /users/:id",
		"POST /login",
		"DELETE /logout",
		"GET /api/v1/users/:id",
		"GET /posts",
		"POST /posts",
		"PATCH /posts",
		"PUT /posts",
		"DELETE /posts",
		"CONNECT /posts",
		"OPTIONS /posts",
		"HEAD /posts",
		"TRACE /posts",
		"POST /register",
	}

	for _, tt := range tests {
		assert.Contains(rs, tt)
	}
}

type StaticRoutingSuiteT struct {
	test.SuiteT
	Recorder *httptest.ResponseRecorder
	Server   *ServerT
}

func (s *StaticRoutingSuiteT) SetupTest() {
	config := support.Config
	config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	s.Recorder = httptest.NewRecorder()
	s.Server = NewServer(config)
}

func (s *StaticRoutingSuiteT) TestStaticRoute() {
	s.Server.Static("/static", "../testdata")
	request, _ := http.NewRequest("GET", "/static/test.html", nil)
	s.Server.ServeHTTP(s.Recorder, request)
	s.Equal(200, s.Recorder.Code)
}

func (s *StaticRoutingSuiteT) TestStaticFSRoute() {
	s.Server.StaticFS("/static_fs", http.Dir("../testdata"))
	request, _ := http.NewRequest("GET", "/static_fs/test.html", nil)
	s.Server.ServeHTTP(s.Recorder, request)
	s.Equal(200, s.Recorder.Code)
}

func (s *StaticRoutingSuiteT) TestStaticFile() {
	s.Server.StaticFile("/static_file/test.html", "../testdata/test.html")
	request, _ := http.NewRequest("GET", "/static_file/test.html", nil)
	s.Server.ServeHTTP(s.Recorder, request)
	s.Equal(200, s.Recorder.Code)
}

func TestStaticRouting(t *testing.T) {
	test.Run(t, new(StaticRoutingSuiteT))
}
