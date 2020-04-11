package record

import (
	"os"
	"testing"

	"github.com/appist/appy/test"
)

type migrationSuite struct {
	test.Suite
}

func (s *migrationSuite) TestModuleName() {
	err := os.Chdir("..")

	s.Nil(err)
	s.Equal("github.com/appist/appy", moduleName())
}

func TestMigrationSuite(t *testing.T) {
	test.Run(t, new(migrationSuite))
}
