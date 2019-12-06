package support

import (
	"bufio"
	"bytes"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	// SugaredLogger is a type alias to zap.SugaredLogger.
	SugaredLogger = zap.SugaredLogger

	// Logger provides the logging functionality.
	Logger struct {
		*SugaredLogger
	}
)

// NewLogger initializes Logger instance.
func NewLogger() *Logger {
	c := newLoggerConfig()
	logger, _ := c.Build()
	defer logger.Sync()

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}
}

// NewFakeLogger initializes a fake Logger instance that is useful for testing purpose.
func NewFakeLogger() (*Logger, *bytes.Buffer, *bufio.Writer) {
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	return &Logger{
		SugaredLogger: zap.New(
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
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
	c.EncoderConfig.EncodeTime = nil

	if IsReleaseBuild() {
		c = zap.NewProductionConfig()
		c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	return c
}
