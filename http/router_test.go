package http

import (
	"testing"

	"github.com/appist/appy/test"
)

type RouterSuite struct {
	test.Suite
}

func (s *RouterSuite) SetupTest() {
}

func (s *RouterSuite) TearDownTest() {
}

func TestRouterSuite(t *testing.T) {
	test.RunSuite(t, new(RouterSuite))
}
