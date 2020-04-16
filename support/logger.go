package support

import (
	"bufio"
	"bytes"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides the logging functionality.
type Logger struct {
	*zap.SugaredLogger
}

// NewLogger initializes Logger instance.
func NewLogger() *Logger {
	c := newLoggerConfig()
	logger, _ := c.Build()
	defer logger.Sync()

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}
}

// NewTestLogger initializes a test Logger instance that is useful for testing purpose.
func NewTestLogger() (*Logger, *bytes.Buffer, *bufio.Writer) {
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)
	c := newLoggerConfig()

	return &Logger{
		SugaredLogger: zap.New(
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(c.EncoderConfig),
				zapcore.AddSync(writer),
				zapcore.DebugLevel,
			),
		).Sugar(),
	}, &buffer, writer
}

func newLoggerConfig() zap.Config {
	c := zap.NewDevelopmentConfig()
	c.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	c.EncoderConfig.CallerKey = ""
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if IsReleaseBuild() {
		c = zap.NewProductionConfig()
	}

	return c
}
