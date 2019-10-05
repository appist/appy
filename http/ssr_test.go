package http

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type SSRSuiteT struct {
	test.SuiteT
	Config     *support.ConfigT
	Recorder   *httptest.ResponseRecorder
	Server     *ServerT
	ViewHelper template.FuncMap
}

func (s *SSRSuiteT) SetupTest() {
	s.Config = support.Config
	s.Config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	s.Server = NewServer(s.Config)
	s.Recorder = httptest.NewRecorder()
	s.ViewHelper = template.FuncMap{
		"testViewHelper": func() string {
			return "i am view helper"
		},
	}

	// Ensure every test case has the correct path
	SSRRootDebug = "app"
	SSRRootRelease = ".ssr"
	ssrRoot = SSRRootDebug
	SSRView = "views"
	SSRLocale = "locales"
	support.Build = "debug"
}

func (s *SSRSuiteT) TestInitSSRWithDebugBuildWithCorrectPath() {
	oldSSRRoot := ssrRoot
	ssrRoot = "../testdata/ssr"
	s.NoError(s.Server.InitSSR(s.ViewHelper))
	s.Server.Router.GET("/", func(c *ContextT) {
		c.HTML(http.StatusOK, "welcome/index", nil)
	})
	request, _ := http.NewRequest("GET", "/", nil)
	s.Server.Router.ServeHTTP(s.Recorder, request)

	s.Equal(500, s.Recorder.Code)

	s.Server = NewServer(s.Config)
	s.NoError(s.Server.InitSSR(s.ViewHelper))
	s.Server.Router.GET("/", func(c *ContextT) {
		c.HTML(http.StatusOK, "welcome/index.html", H{"message": "i am testing"})
	})
	s.Recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/", nil)
	s.Server.Router.ServeHTTP(s.Recorder, request)

	s.Equal(200, s.Recorder.Code)
	s.Contains(s.Recorder.Body.String(), "i am testing")
	s.Contains(s.Recorder.Body.String(), "i am view helper")
	ssrRoot = oldSSRRoot
}

func (s *SSRSuiteT) TestInitSSRWithDebugBuildWithIncorrectPath() {
	ssrRoot = "../testdata/ssr_only_views"
	s.EqualError(s.Server.InitSSR(s.ViewHelper), fmt.Sprintf("open %s/locales: no such file or directory", ssrRoot))

	ssrRoot = "../testdata/ssr_only_locales"
	s.EqualError(s.Server.InitSSR(s.ViewHelper), fmt.Sprintf("open %s/views: no such file or directory", ssrRoot))
}

func (s *SSRSuiteT) TestInitSSRWithReleaseBuildWithCorrectPath() {
	support.Build = "release"
	SSRRootRelease = "."
	s.Server.Assets = http.Dir("../testdata/ssr")
	s.NoError(s.Server.InitSSR(s.ViewHelper))
	s.Server.Router.GET("/", func(c *ContextT) {
		c.HTML(http.StatusOK, "welcome/index", nil)
	})
	request, _ := http.NewRequest("GET", "/", nil)
	s.Server.Router.ServeHTTP(s.Recorder, request)

	s.Equal(500, s.Recorder.Code)

	s.Server = NewServer(s.Config)
	s.Server.Assets = http.Dir("../testdata/ssr")
	s.NoError(s.Server.InitSSR(s.ViewHelper))
	s.Server.Router.GET("/", func(c *ContextT) {
		c.HTML(http.StatusOK, "welcome/index.html", H{"message": "i am testing"})
	})
	s.Recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/", nil)
	s.Server.Router.ServeHTTP(s.Recorder, request)

	s.Equal(200, s.Recorder.Code)
	s.Contains(s.Recorder.Body.String(), "i am testing")
	s.Contains(s.Recorder.Body.String(), "i am view helper")
}

func (s *SSRSuiteT) TestInitSSRWithReleaseBuildWithIncorrectPath() {
	support.Build = "release"
	assets := "../testdata/assets/.ssr_only_views"
	s.Server.Assets = http.Dir(assets)
	s.EqualError(s.Server.InitSSR(s.ViewHelper), fmt.Sprintf("open %s/.ssr/locales: no such file or directory", assets))

	assets = "../testdata/assets/.ssr_only_locales"
	s.Server.Assets = http.Dir(assets)
	s.EqualError(s.Server.InitSSR(s.ViewHelper), fmt.Sprintf("open %s/.ssr/views: no such file or directory", assets))
}

func TestSSR(t *testing.T) {
	test.Run(t, new(SSRSuiteT))
}
