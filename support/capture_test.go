package support

import (
	"fmt"
	"os"
	"testing"

	"github.com/appist/appy/test"
)

func TestCaptureLogOutput(t *testing.T) {
	assert := test.NewAssert(t)
	output := CaptureLogOutput(func() {
		Logger.Debug("test debug")
		Logger.Error("test error")
		Logger.Info("test info")
		Logger.Warn("test warn")
	})

	assert.Contains(output, "DEBUG\ttest debug")
	assert.Contains(output, "ERROR\ttest error")
	assert.Contains(output, "INFO\ttest info")
	assert.Contains(output, "WARN\ttest warn")
}

func TestCaptureOutput(t *testing.T) {
	assert := test.NewAssert(t)
	output := CaptureOutput(func() {
		fmt.Fprint(os.Stdout, "foo")
		fmt.Fprint(os.Stderr, "bar")
	})

	assert.Equal("foobar", output)
}
