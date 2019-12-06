package support

import (
	"testing"

	"github.com/appist/appy/test"
)

type SupportSuite struct {
	test.Suite
}

func (s *SupportSuite) SetupTest() {
}

func (s *SupportSuite) TearDownTest() {
}

func (s *SupportSuite) TestIsDebugBuild() {
	s.Equal(true, IsDebugBuild())
}

func (s *SupportSuite) TestIsReleaseBuild() {
	s.Equal(false, IsReleaseBuild())
}

func TestSupportSuite(t *testing.T) {
	test.RunSuite(t, new(SupportSuite))
}
