package appy

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

func (s *CaptureSuite) TestCaptureLoggerOutput() {
	output := CaptureLoggerOutput(func() {
		Logger.Debug("test debug")
		Logger.Error("test error")
		Logger.Info("test info")
		Logger.Warn("test warn")
	})

	s.Contains(output, "DEBUG\ttest debug")
	s.Contains(output, "ERROR\ttest error")
	s.Contains(output, "INFO\ttest info")
	s.Contains(output, "WARN\ttest warn")
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
