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
	support.Init(nil)
	s.Config = &support.ConfigT{}
	support.DeepClone(&s.Config, &support.Config)
	s.Config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	s.Server = NewServer(s.Config)
	spaResources = []spaResourceT{}
}

func (s *InitCSRSuiteT) TestAssetsNotConfigured() {
	s.Server.InitCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/login", nil)
	s.Server.Router.ServeHTTP(recorder, request)

	s.Equal(404, recorder.Code)
	support.Logger.Info(recorder.Body.String())
}

func (s *InitCSRSuiteT) TestNonExistingPathWithAssetsNil() {
	s.Server.Assets = http.Dir("../testdata")
	s.Server.InitCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/login", nil)
	s.Server.Router.ServeHTTP(recorder, request)

	s.Equal(404, recorder.Code)
	support.Logger.Info(recorder.Body.String())
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

func (s *InitCSRSuiteT) TestToolsSPA() {
	s.Server.Assets = http.Dir("../testdata/assets")
	s.Server.InitCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/tools", nil)
	s.Server.Router.ServeHTTP(recorder, request)

	s.Equal(200, recorder.Code)
	s.Contains(recorder.Body.String(), "<title>appy tooling</title>")
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
