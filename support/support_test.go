package support

import (
	"testing"

	"github.com/appist/appy/test"
)

type supportSuite struct {
	test.Suite
}

func (s *supportSuite) TestIsDebugBuild() {
	Build = DebugBuild

	s.Equal(true, IsDebugBuild())
	s.Equal(false, IsReleaseBuild())
}

func (s *supportSuite) TestIsReleaseBuild() {
	Build = ReleaseBuild
	defer func() { Build = DebugBuild }()

	s.Equal(false, IsDebugBuild())
	s.Equal(true, IsReleaseBuild())
}

func TestSupportSuite(t *testing.T) {
	test.Run(t, new(supportSuite))
}
