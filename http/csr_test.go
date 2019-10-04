package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type InitCSRSuiteT struct {
	test.SuiteT
	Config *support.ConfigT
	Server *ServerT
}

func (s *InitCSRSuiteT) SetupTest() {
	s.Config = &support.ConfigT{}
	support.Copy(&s.Config, &support.Config)
	s.Config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	s.Server = NewServer(s.Config)
}

func (s *InitCSRSuiteT) TestAssetsNotConfigured() {
	s.Server.InitCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/login", nil)
	s.Server.Router.ServeHTTP(recorder, request)

	s.Equal(404, recorder.Code)
}

func (s *InitCSRSuiteT) TestNonExistingPathWithAssetsNil() {
	s.Server.Assets = http.Dir("../testdata")
	s.Server.InitCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/login", nil)
	s.Server.Router.ServeHTTP(recorder, request)

	s.Equal(404, recorder.Code)
}

func (s *InitCSRSuiteT) TestNonExistingPathWithAssetsNotNil() {
	s.Server.Assets = http.Dir("../testdata/assets")
	s.Server.InitCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/login", nil)
	s.Server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Contains(recorder.Body.String(), "<div id=\"app\">we build apps</div>")
}

func (s *InitCSRSuiteT) TestStaticAssets301Redirect() {
	s.Server.Assets = http.Dir("../testdata/assets")
	s.Server.InitCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/index.html", nil)
	s.Server.Router.ServeHTTP(recorder, request)

	s.Equal(301, recorder.Code)
}

func (s *InitCSRSuiteT) TestFallbackReturnsIndexHTML() {
	s.Server.Assets = http.Dir("../testdata/assets")
	s.Server.InitCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/login", nil)
	s.Server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Contains(recorder.Body.String(), "<div id=\"app\">we build apps</div>")
}

// We need to better test this in the integration test.
func (s *InitCSRSuiteT) TestSEOCrawlingWithSSLDisabled() {
	s.Server.Assets = http.Dir("../testdata/assets")
	s.Server.InitCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/", nil)
	request.Host = "0.0.0.0"
	request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	s.Server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Equal("text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Equal("<html><head></head><body></body></html>", recorder.Body.String())
}

// We need to better test this in the integration test.
func (s *InitCSRSuiteT) TestCrawlerBotWithSSLEnabled() {
	s.Server.Assets = http.Dir("../testdata/assets")
	s.Server.InitCSR()
	s.Server.Config.HTTPSSLEnabled = true

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/", nil)
	request.Host = "https"
	request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	s.Server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Equal("text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Equal("<html><head></head><body></body></html>", recorder.Body.String())
}

func (s *InitCSRSuiteT) TestCrawlerBotWithMalformedURL() {
	s.Server.Assets = http.Dir("../testdata/assets")
	s.Server.InitCSR()
	s.Server.Config.HTTPSSLEnabled = true

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "https://blabla", nil)
	request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)")
	s.Server.Router.ServeHTTP(recorder, request)

	s.Equal(500, recorder.Code)
}

func (s *InitCSRSuiteT) TestSSRStillWorkCorrectly() {
	s.Server.Assets = http.Dir("../testdata/assets")
	s.Server.InitCSR()
	s.Server.Router.GET("/welcome", func(c *ContextT) {
		c.String(200, "%s", "test")
	})

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/welcome", nil)
	s.Server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Equal("test", recorder.Body.String())
}

func TestInitCSR(t *testing.T) {
	test.Run(t, new(InitCSRSuiteT))
}
