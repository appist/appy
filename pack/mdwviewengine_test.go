package pack

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/appist/appy/support"
	"github.com/appist/appy/test"
	"github.com/appist/appy/view"
)

type mdwViewEngineSuite struct {
	test.Suite
	asset      *support.Asset
	config     *support.Config
	logger     *support.Logger
	viewEngine *view.Engine
}

func (s *mdwViewEngineSuite) SetupTest() {
	os.Setenv("APPY_ENV", "development")
	os.Setenv("APPY_MASTER_KEY", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")

	s.logger, _, _ = support.NewTestLogger()
	s.asset = support.NewAsset(nil, "")
	s.config = support.NewConfig(s.asset, s.logger)
	s.viewEngine = view.NewEngine(s.asset, s.config, s.logger)
}

func (s *mdwViewEngineSuite) TearDownTest() {
	os.Unsetenv("APPY_ENV")
	os.Unsetenv("APPY_MASTER_KEY")
	os.Unsetenv("HTTP_CSRF_SECRET")
	os.Unsetenv("HTTP_SESSION_SECRETS")
}

func (s *mdwViewEngineSuite) TestExistence() {
	c, _ := NewTestContext(httptest.NewRecorder())
	mdwViewEngine(s.asset, s.config, s.logger, nil)(c)

	s.NotNil(c.Get(mdwViewEngineCtxKey.String()))
}

func TestMdwViewEngineSuite(t *testing.T) {
	test.Run(t, new(mdwViewEngineSuite))
}
