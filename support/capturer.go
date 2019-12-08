package support

import (
	"bytes"
	"io"
	"os"
)

type (
	capturer struct {
		stdout bool
		stderr bool
	}
)

func (c *capturer) capture(f func()) (string, error) {
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
	_, err := io.Copy(&buf, r)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// CaptureOutput captures stdout and stderr.
func CaptureOutput(f func()) (string, error) {
	capturer := &capturer{stdout: true, stderr: true}
	return capturer.capture(f)
}
