package appy_test

import (
	"testing"

	"github.com/appist/appy"
)

type AppySuite struct {
	appy.TestSuite
}

func (s *AppySuite) TestIsDebugBuild() {
	s.Equal(true, appy.IsDebugBuild())
}

func (s *AppySuite) TestIsReleaseBuild() {
	s.Equal(false, appy.IsReleaseBuild())
}

func TestAppySuite(t *testing.T) {
	appy.RunTestSuite(t, new(AppySuite))
}
