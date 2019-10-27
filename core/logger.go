package core

import (
	"context"
	"errors"
	"time"

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

// BeforeQuery is a hook before a go-pg query.
func (l AppLogger) BeforeQuery(c context.Context, q *AppDbQueryEvent) (context.Context, error) {
	return c, nil
}

// AfterQuery is a hook after a go-pg query.
func (l AppLogger) AfterQuery(c context.Context, q *AppDbQueryEvent) error {
	query, err := q.FormattedQuery()
	if err != nil {
		return err
	}

	l.SugaredLogger.Infof("[DB] %s in %s", query, time.Since(q.StartTime))
	return nil
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
