package appy

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type AttachViewEngineSuite struct {
	TestSuite
	asset      *Asset
	config     *Config
	logger     *Logger
	viewEngine *ViewEngine
}

func (s *AttachViewEngineSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, _, _ = NewFakeLogger()
	layout := map[string]string{
		"docker": "testdata/.docker",
		"config": "testdata/configs",
		"locale": "testdata/pkg/locales",
		"view":   "testdata/pkg/views",
		"web":    "testdata/web",
	}
	s.asset = NewAsset(http.Dir("testdata"), layout, "")
	s.config = NewConfig(s.asset, s.logger, &Support{})
	s.viewEngine = NewViewEngine(s.asset, s.config, s.logger)
}

func (s *AttachViewEngineSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *AttachViewEngineSuite) TestExistence() {
	c, _ := NewTestContext(httptest.NewRecorder())
	AttachViewEngine(s.asset, s.config, s.logger, nil)(c)
	s.NotNil(c.Get(viewEngineCtxKey.String()))
}

func TestAttachViewEngineSuite(t *testing.T) {
	RunTestSuite(t, new(AttachViewEngineSuite))
}
