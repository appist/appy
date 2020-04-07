package support

import (
	"bytes"
	"io"
	"os"
)

type capturer struct {
	stdout bool
	stderr bool
}

func (c *capturer) capture(f func()) string {
	r, w, _ := os.Pipe()

	if c.stdout {
		stdout := os.Stdout
		os.Stdout = w
		defer func() {
			os.Stdout = stdout
		}()
	}

	if c.stderr {
		stderr := os.Stderr
		os.Stderr = w
		defer func() {
			os.Stderr = stderr
		}()
	}

	f()
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String()
}

// CaptureOutput captures stdout and stderr.
func CaptureOutput(f func()) string {
	capturer := &capturer{
		stdout: true,
		stderr: true,
	}

	return capturer.capture(f)
}
