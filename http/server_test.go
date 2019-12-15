package http

import (
	"testing"

	"github.com/appist/appy/test"
)

type ServerSuite struct {
	test.Suite
}

func (s *ServerSuite) SetupTest() {
}

func (s *ServerSuite) TearDownTest() {
}

func TestServerSuite(t *testing.T) {
	test.RunSuite(t, new(ServerSuite))
}
