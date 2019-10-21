package core

import (
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerConfig = zap.Config

// SugaredLogger is a type alias to zap.SugaredLogger.
type SugaredLogger = zap.SugaredLogger

// AppLogger keeps the logging functionality.
type AppLogger struct {
	*SugaredLogger
}

func newLoggerConfig() loggerConfig {
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

func newLogger(c loggerConfig) (*AppLogger, error) {
	logger, err := c.Build()
	if err != nil {
		return nil, errors.New("unable to build the logger")
	}

	defer logger.Sync()
	return &AppLogger{
		SugaredLogger: logger.Sugar(),
	}, nil
}
