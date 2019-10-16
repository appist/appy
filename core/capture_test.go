package core

import (
	"fmt"
	"os"
	"testing"

	"github.com/appist/appy/test"
)

type CaptureSuite struct {
	test.Suite
}

func (s *CaptureSuite) SetupTest() {
}

func (s *CaptureSuite) TearDownTest() {
}

func (s *CaptureSuite) TestCaptureOutput() {
	output := CaptureOutput(func() {
		fmt.Fprint(os.Stdout, "foo")
		fmt.Fprint(os.Stderr, "bar")
	})

	s.Equal("foobar", output)
}

func TestCapture(t *testing.T) {
	test.Run(t, new(CaptureSuite))
}
