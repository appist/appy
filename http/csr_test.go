package http

import (
	"appist/appy/support"
	"appist/appy/test"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type SetupCSRSuiteT struct {
	test.SuiteT
	Server *ServerT
}

func (s *SetupCSRSuiteT) SetupTest() {
	config := support.Config
	config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	s.Server = NewServer(config)
}

func (s *SetupCSRSuiteT) TestAssetsNotConfigured() {
	s.Server.SetupCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/login", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(404, recorder.Code)
}

func (s *SetupCSRSuiteT) TestIndexHTMLMissing() {
	s.Server.SetupAssets(http.Dir("."))
	s.Server.SetupCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/login", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(404, recorder.Code)
}

func (s *SetupCSRSuiteT) TestNonExistingRequest() {
	s.Server.SetupAssets(http.Dir("../testdata"))
	s.Server.SetupCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/login", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(404, recorder.Code)
}

func (s *SetupCSRSuiteT) TestStaticAssets301Redirect() {
	s.Server.SetupAssets(http.Dir("../testdata"))
	s.Server.SetupCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/index.html", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(301, recorder.Code)
}

func (s *SetupCSRSuiteT) TestFallbackReturnsIndexHTML() {
	s.Server.SetupAssets(http.Dir("../testdata"))
	s.Server.SetupCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/login", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
}

func (s *SetupCSRSuiteT) TestHTTPCrawlerBotWithSSLDisabled() {
	s.Server.SetupAssets(http.Dir("../testdata"))
	s.Server.SetupCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/login", nil)
	request.Host = "0.0.0.0"
	request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
}

func (s *SetupCSRSuiteT) TestCrawlerBotWithSSLEnabled() {
	s.Server.SetupAssets(http.Dir("../testdata"))
	s.Server.SetupCSR()
	s.Server.Config.HTTPSSLEnabled = true

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/login", nil)
	request.Host = "0.0.0.0"
	request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
}

func (s *SetupCSRSuiteT) TestSSRWorksCorrectly() {
	s.Server.SetupAssets(http.Dir("../testdata"))
	s.Server.GET("/welcome", func(c *gin.Context) {
		s.Server.RenderString(c, 200, "%s", "test")
	})
	s.Server.SetupCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/welcome", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(200, recorder.Code)
	s.Equal("test", recorder.Body.String())
}

func TestSetupCSR(t *testing.T) {
	test.Run(t, new(SetupCSRSuiteT))
}
