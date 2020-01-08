package appy_test

import (
	"testing"

	"github.com/appist/appy"
)

type SupportSuite struct {
	appy.TestSuite
}

func (s *SupportSuite) TestIsDebugBuild() {
	s.Equal(true, appy.IsDebugBuild())
}

func (s *SupportSuite) TestIsReleaseBuild() {
	s.Equal(false, appy.IsReleaseBuild())
}

func TestSupportSuite(t *testing.T) {
	appy.RunTestSuite(t, new(SupportSuite))
}
