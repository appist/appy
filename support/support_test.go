package support

import (
	"testing"

	"github.com/appist/appy/test"
)

type SupportSuite struct {
	test.Suite
}

func (s *SupportSuite) TestIsDebugBuild() {
	Build = DebugBuild

	s.Equal(true, IsDebugBuild())
	s.Equal(false, IsReleaseBuild())
}

func (s *SupportSuite) TestIsReleaseBuild() {
	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()

	s.Equal(false, IsDebugBuild())
	s.Equal(true, IsReleaseBuild())
}

func TestSupportSuite(t *testing.T) {
	test.Run(t, new(SupportSuite))
}
