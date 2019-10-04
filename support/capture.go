package support

import (
	"bufio"
	"bytes"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// CaptureOutput captures the logger's output.
func CaptureOutput(f func()) string {
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
