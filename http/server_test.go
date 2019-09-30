package http

import (
	"appist/appy/support"
	"appist/appy/test"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewServer(t *testing.T) {
	s := NewServer(support.Config)
	assert := test.NewAssert(t)
	assert.NotEqual(nil, s.router)
	assert.NotEqual(nil, s.htmlRenderer)
	assert.NotEqual(nil, s.Config)
	assert.NotEqual(nil, s.HTTP)
}

func TestGetAllRoutes(t *testing.T) {
	s := NewServer(support.Config)
	assert := test.NewAssert(t)
	assert.Equal(1, len(s.GetAllRoutes()))
	s.GET("/render", func(c *gin.Context) {
		s.RenderJSON(c, http.StatusOK, gin.H{"a": 1})
	})
	assert.Equal(2, len(s.GetAllRoutes()))
}

func TestSecureJSONPrefix(t *testing.T) {
	s := NewServer(support.Config)
	s.GET("/render", func(c *gin.Context) {
		s.RenderSecureJSON(c, http.StatusOK, []string{"lena", "austin", "foo"})
	})
	s.SecureJSONPrefix(";;")
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.ServeHTTP(recorder, request)

	assert := test.NewAssert(t)
	assert.Equal(200, recorder.Code)
	assert.Equal("application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
	assert.Equal(";;[\"lena\",\"austin\",\"foo\"]", recorder.Body.String())
}

func TestSetupAssets(t *testing.T) {
	s := NewServer(support.Config)
	s.SetupAssets(http.Dir("."))

	assert := test.NewAssert(t)
	assert.NotEqual(nil, s.assets)
}

func TestSetFuncMap(t *testing.T) {
	oldViewDir := ViewDir
	ViewDir = "."
	s := NewServer(support.Config)
	s.SetupAssets(http.Dir("../testdata"))
	s.SetFuncMap(template.FuncMap{
		"date": func() string {
			return "2019-01-01"
		},
	})
	s.AddView("test", "layout.html", []string{"viewHelper.html"})
	s.GET("/render", func(c *gin.Context) {
		s.RenderHTML(c, http.StatusOK, "test", nil)
	})
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/render", nil)
	s.ServeHTTP(recorder, request)

	assert := test.NewAssert(t)
	assert.Equal(200, recorder.Code)
	assert.Equal("text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	assert.Equal("<div>Date: 2019-01-01</div>", strings.Trim(recorder.Body.String(), "\n"))
	ViewDir = oldViewDir
}
