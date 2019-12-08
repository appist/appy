package support

import (
	"fmt"
	"os"
	"testing"

	"github.com/appist/appy/test"
)

type CapturerSuite struct {
	test.Suite
}

func (s *CapturerSuite) SetupTest() {
}

func (s *CapturerSuite) TearDownTest() {
}

func (s *CapturerSuite) TestCaptureOutput() {
	output, err := CaptureOutput(func() {
		fmt.Fprint(os.Stdout, "foo")
		fmt.Fprint(os.Stderr, "bar")
	})

	s.Equal("foobar", output)
	s.Nil(err)
}

func TestCapturerSuite(t *testing.T) {
	test.RunSuite(t, new(CapturerSuite))
}
