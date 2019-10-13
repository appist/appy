package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type ServerSuiteT struct {
	test.SuiteT
	Config *support.ConfigT
	Server *ServerT
}

func (s *ServerSuiteT) SetupTest() {
	support.Init(nil)
	s.Config = &support.ConfigT{}
	support.Copy(&s.Config, &support.Config)
	s.Config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	s.Server = NewServer(s.Config)
}

func (s *ServerSuiteT) TestNewServerWithoutSSLEnabled() {
	s.Server.Assets = http.Dir("../testdata/assets")
	s.NotNil(s.Server.Assets)
	s.NotNil(s.Server.Config)
	s.NotNil(s.Server.HTTP)
	s.NotNil(s.Server.HTMLRenderer)
	s.NotNil(s.Server.Router)
	s.Equal("0.0.0.0:3000", s.Server.HTTP.Addr)
}

func (s *ServerSuiteT) TestNewServerWithSSLEnabled() {
	s.Config.HTTPSSLEnabled = true
	s.Server = NewServer(s.Config)
	s.Server.Assets = http.Dir("../testdata/assets")
	s.NotNil(s.Server.Assets)
	s.NotNil(s.Server.Config)
	s.NotNil(s.Server.HTTP)
	s.NotNil(s.Server.HTMLRenderer)
	s.NotNil(s.Server.Router)
	s.Equal("0.0.0.0:3443", s.Server.HTTP.Addr)
}

func (s *ServerSuiteT) TestAddDefaultWelcomePage() {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/", nil)
	s.Server = NewServer(s.Config)
	s.Server.AddDefaultWelcomePage()
	s.Server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Equal("text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Contains(recorder.Body.String(), "<p class=\"lead\">An opinionated productive web framework that helps scaling business easier.</p>")

	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/", nil)
	s.Server = NewServer(s.Config)
	s.Server.Router.GET("/", func(c *ContextT) {
		c.JSON(200, H{"a": 1})
	})
	s.Server.AddDefaultWelcomePage()
	s.Server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Equal("application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Equal("{\"a\":1}\n", recorder.Body.String())

	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/", nil)
	s.Server = NewServer(s.Config)
	s.Server.Assets = http.Dir("../testdata/assets")
	s.Server.InitCSR()
	s.Server.AddDefaultWelcomePage()
	s.Server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Equal("text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
	s.Contains(recorder.Body.String(), "we build apps")
}

func (s *ServerSuiteT) TestIsSSLCertsExist() {
	s.Equal(false, s.Server.IsSSLCertsExist())

	s.Config.HTTPSSLCertPath = "../testdata/ssl"
	s.Equal(true, s.Server.IsSSLCertsExist())
}

func (s *ServerSuiteT) TestDebugPrintInfo() {
	output := support.CaptureOutput(func() {
		s.Server.PrintInfo()
	})

	s.Contains(output, fmt.Sprintf("* Version 0.1.0 (%s), build: debug, environment: development, config: none", runtime.Version()))
	s.Contains(output, "* Listening on http://0.0.0.0:3000")

	s.Config.HTTPSSLEnabled = true
	output = support.CaptureOutput(func() {
		s.Server.PrintInfo()
	})

	s.Contains(output, fmt.Sprintf("* Version 0.1.0 (%s), build: debug, environment: development, config: none", runtime.Version()))
	s.Contains(output, "* Listening on https://0.0.0.0:3443")
}

func (s *ServerSuiteT) TestReleasePrintInfo() {
	oldBuild := support.Build
	support.Build = "release"
	output := support.CaptureLogOutput(func() {
		s.Server.PrintInfo()
	})

	s.Contains(output, fmt.Sprintf("* Version 0.1.0 (%s), build: release, environment: development, config: none", runtime.Version()))
	s.Contains(output, "* Listening on http://0.0.0.0:3000")

	s.Config.HTTPSSLEnabled = true
	output = support.CaptureLogOutput(func() {
		s.Server.PrintInfo()
	})

	s.Contains(output, fmt.Sprintf("* Version 0.1.0 (%s), build: release, environment: development, config: none", runtime.Version()))
	s.Contains(output, "* Listening on https://0.0.0.0:3443")
	support.Build = oldBuild
}

func TestServer(t *testing.T) {
	test.Run(t, new(ServerSuiteT))
}
