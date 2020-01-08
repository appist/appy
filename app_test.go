package appy_test

import (
	"testing"

	"github.com/appist/appy"
)

type AppSuite struct {
	appy.TestSuite
}

func (s *AppSuite) TestNewApp() {
	app := appy.NewApp()

	s.NotNil(app.Logger())
}

func TestAppSuite(t *testing.T) {
	appy.RunTestSuite(t, new(AppSuite))
}
