package support

import (
	"os"
	"testing"

	"github.com/appist/appy/test"
)

type moduleSuite struct {
	test.Suite
}

func (s *moduleSuite) TestModuleName() {
	s.Equal("", ModuleName())

	err := os.Chdir("..")
	s.Nil(err)
	s.Equal("github.com/appist/appy", ModuleName())
}

func TestModuleSuite(t *testing.T) {
	test.Run(t, new(moduleSuite))
}
