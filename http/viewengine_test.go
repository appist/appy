package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
)

type ViewEngineSuite struct {
	test.Suite
	assets     *support.Assets
	config     *support.Config
	logger     *support.Logger
	viewEngine *support.ViewEngine
}

func (s *ViewEngineSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, _, _ = support.NewFakeLogger()
	layout := map[string]string{
		"docker": "../support/testdata/.docker",
		"config": "../support/testdata/configs",
		"locale": "../support/testdata/pkg/locales",
		"view":   "../support/testdata/pkg/views",
		"web":    "../support/testdata/web",
	}
	s.assets = support.NewAssets(layout, "", http.Dir("../support/testdata"))
	s.config = support.NewConfig(s.assets, s.logger)
	s.viewEngine = support.NewViewEngine(s.assets)
}

func (s *ViewEngineSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *ViewEngineSuite) TestExistence() {
	c, _ := NewTestContext(httptest.NewRecorder())
	ViewEngine(s.assets, nil)(c)
	s.NotNil(c.Get(viewEngineCtxKey.String()))
}

func TestViewEngineSuite(t *testing.T) {
	test.RunSuite(t, new(ViewEngineSuite))
}
