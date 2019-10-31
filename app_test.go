package appy_test

import (
	"testing"

	"github.com/appist/appy"
)

type AppSuite struct {
	appy.TestSuite
	app appy.App
}

func (s *AppSuite) SetupTest() {
	s.app = appy.NewApp()
}

func (s *AppSuite) TearDownTest() {
}

func (s *AppSuite) TestAppCmd() {
}

func (s *AppSuite) TestAppConfig() {
}

func (s *AppSuite) TestAppDbManager() {
}

func (s *AppSuite) TestAppLogger() {
	s.NotNil(s.app.Logger())
}

func (s *AppSuite) TestAppServer() {
}

func (s *AppSuite) TestAppSupport() {
	s.NotNil(s.app.Support())
}

func TestApp(t *testing.T) {
	appy.RunTestSuite(t, new(AppSuite))
}
