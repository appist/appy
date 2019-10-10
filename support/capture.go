package support

import (
	"bufio"
	"bytes"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type capturer struct {
	stdout bool
	stderr bool
}

// CaptureLogOutput captures the logger's output.
func CaptureLogOutput(f func()) string {
	var buffer bytes.Buffer
	oldLogger := Logger
	encoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	writer := bufio.NewWriter(&buffer)
	Logger = zap.New(zapcore.NewCore(encoder, zapcore.AddSync(writer), zapcore.DebugLevel)).Sugar()
	f()
	writer.Flush()
	Logger = oldLogger

	return buffer.String()
}

// CaptureOutput captures stdout and stderr.
func CaptureOutput(f func()) string {
	capturer := &capturer{stdout: true, stderr: true}
	return capturer.capture(f)
}

func (c *capturer) capture(f func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

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
