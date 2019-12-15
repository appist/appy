package http

import (
	"testing"

	"github.com/appist/appy/test"
)

type ContextSuite struct {
	test.Suite
}

func (s *ContextSuite) SetupTest() {
}

func (s *ContextSuite) TearDownTest() {
}

func (s *ContextSuite) TestArrayContains() {
}

func TestContextSuite(t *testing.T) {
	test.RunSuite(t, new(ContextSuite))
}
