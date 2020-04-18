package record

import (
	"testing"

	"github.com/appist/appy/test"
)

type migrationSuite struct {
	test.Suite
}

func TestMigrationSuite(t *testing.T) {
	test.Run(t, new(migrationSuite))
}
