package http

import (
	"appist/appy/support"
	"appist/appy/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

type SetupCSRSuite struct {
	test.SuiteT
	Config *support.ConfigT
	Server *ServerT
}

func (s *SetupCSRSuite) SetupTest() {
	s.Config = support.Config
	s.Server = NewServer(s.Config)
}

func (s *SetupCSRSuite) TearDownAllTest() {
}

func (s *SetupCSRSuite) TearDownTest() {
}

func (s *SetupCSRSuite) TestAssetsNotConfigured() {
	s.Server.SetupCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/login", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(404, recorder.Code)
}

func (s *SetupCSRSuite) TestIndexHTMLMissing() {
	s.Server.SetupAssets(http.Dir("."))
	s.Server.SetupCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/login", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(404, recorder.Code)
}

func (s *SetupCSRSuite) TestNonExistingRequest() {
	s.Server.SetupAssets(http.Dir("../testdata"))
	s.Server.SetupCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/login", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(404, recorder.Code)
}

func (s *SetupCSRSuite) TestStaticAssets301Redirect() {
	s.Server.SetupAssets(http.Dir("../testdata"))
	s.Server.SetupCSR()

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/index.html", nil)
	s.Server.ServeHTTP(recorder, request)
	s.Equal(301, recorder.Code)
}

func TestSetupCSR(t *testing.T) {
	test.Run(t, new(SetupCSRSuite))
}
