package http

import (
	"appist/appy/support"
	"appist/appy/test"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type SSRSuiteT struct {
	test.SuiteT
	Recorder   *httptest.ResponseRecorder
	Server     *ServerT
	OldViewDir string
}

func (s *SSRSuiteT) SetupTest() {
	s.OldViewDir = ViewDir
	ViewDir = "."
	config := support.Config
	config.HTTPCSRFSecret = []byte("481e5d98a31585148b8b1dfb6a3c0465")
	s.Server = NewServer(config)
	s.Server.SetupAssets(http.Dir("../testdata"))
	s.Recorder = httptest.NewRecorder()
}

func (s *SSRSuiteT) TearDownTest() {
	ViewDir = s.OldViewDir
}

func (s *SSRSuiteT) TestAddViewRendersCorrectly() {
	s.Server.GET("/", func() gin.HandlerFunc {
		s.Server.AddView(
			"test",
			"layout.html",
			[]string{"content.html"},
		)

		return func(c *gin.Context) {
			c.HTML(http.StatusOK, "test", nil)
		}
	}())

	request, _ := http.NewRequest("GET", "/", nil)
	s.Server.ServeHTTP(s.Recorder, request)
	s.Equal(200, s.Recorder.Code)
	s.Contains(s.Recorder.Body.String(), "i am content")
}

func (s *SSRSuiteT) TestAddViewReturnsError() {
	err := s.Server.AddView(
		"test",
		"missing.html",
		[]string{"content.html"},
	)

	s.NotNil(err)

	err = s.Server.AddView(
		"test",
		"layout.html",
		[]string{"missing.html"},
	)

	s.NotNil(err)
}

func (s *SSRSuiteT) TestSetupI18n() {
	err := s.Server.SetupI18n()
	s.Nil(err)

	s.Server.GET("/", func() gin.HandlerFunc {
		s.Server.AddView(
			"test",
			"layout.html",
			[]string{"i18n.html"},
		)

		return func(c *gin.Context) {
			s.Server.RenderHTML(c, http.StatusOK, "test", map[string]interface{}{
				"Messages": map[string]interface{}{
					"Name":  "test",
					"Count": 1,
				},
			})
		}
	}())

	request, _ := http.NewRequest("GET", "/", nil)
	s.Server.ServeHTTP(s.Recorder, request)
	s.Equal(200, s.Recorder.Code)
	s.Contains(s.Recorder.Body.String(), "test has 1 message.")

	oldLocaleDir := LocaleDir
	LocaleDir = "../testdata"
	err = s.Server.SetupI18n()
	s.NotNil(err)
	LocaleDir = oldLocaleDir
}

func TestSSR(t *testing.T) {
	test.Run(t, new(SSRSuiteT))
}
