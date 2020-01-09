package appy_test

import (
	"testing"

	"github.com/appist/appy"
)

type AppSuite struct {
	appy.TestSuite
}

func (s *AppSuite) TestNewApp() {
	asset := appy.NewAsset(nil, map[string]string{
		"docker": "testdata/app/.docker",
		"config": "testdata/app/configs",
		"locale": "testdata/app/pkg/locales",
		"view":   "testdata/app/pkg/views",
		"web":    "testdata/app/web",
	})
	app := appy.NewApp(asset)

	s.NotNil(app.Asset())
	s.NotNil(app.Config())
	s.NotNil(app.I18n())
	s.NotNil(app.Logger())
}

func TestAppSuite(t *testing.T) {
	appy.RunTestSuite(t, new(AppSuite))
}
