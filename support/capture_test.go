package support

import (
	"fmt"
	"os"
	"testing"

	"github.com/appist/appy/test"
)

type captureSuite struct {
	test.Suite
}

func (s *captureSuite) TestCaptureOutput() {
	output := CaptureOutput(func() {
		fmt.Fprint(os.Stdout, "foo")
		fmt.Fprint(os.Stderr, "bar")
	})

	s.Equal("foobar", output)
}

func TestCaptureSuite(t *testing.T) {
	test.Run(t, new(captureSuite))
}
