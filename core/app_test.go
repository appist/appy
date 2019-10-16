package core

import (
	"os"
	"testing"

	"github.com/appist/appy/test"
)

type AppSuite struct {
	test.Suite
}

func (s *AppSuite) SetupTest() {
}

func (s *AppSuite) TearDownTest() {
	os.Clearenv()
}

func (s *AppSuite) TestNewApp() {
	os.Setenv("HTTP_CSRF_SECRET", "481e5d98a31585148b8b1dfb6a3c0465")
	os.Setenv("HTTP_SESSION_SECRETS", "481e5d98a31585148b8b1dfb6a3c0465")
	app, err := NewApp(nil, nil)
	s.Nil(err)
	s.NotNil(app.Config)
	s.NotNil(app.Logger)
	s.NotNil(app.Server)
}

func TestApp(t *testing.T) {
	test.Run(t, new(AppSuite))
}
