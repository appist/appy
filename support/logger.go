package support

import (
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerT is a type alias to zap.SugaredLogger.
type LoggerT = zap.SugaredLogger

// NewLoggerConfig returns logger configuration.
func NewLoggerConfig() zap.Config {
	c := zap.NewDevelopmentConfig()
	c.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	c.EncoderConfig.CallerKey = ""
	c.EncoderConfig.EncodeTime = nil

	if Build != "debug" {
		c = zap.NewProductionConfig()
		c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	return c
}

// NewLogger initiates a logger instance.
func NewLogger(c zap.Config) (*LoggerT, error) {
	logger, err := c.Build()
	if err != nil {
		return nil, errors.New("unable to build the logger")
	}

	defer logger.Sync()
	return logger.Sugar(), nil
}
