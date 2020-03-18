package appy

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	// Logger provides the logging functionality.
	Logger struct {
		*zap.SugaredLogger
		dbLogging bool
	}
)

const (
	dbQueryComment = "/* appy framework */"
)

// NewLogger initializes Logger instance.
func NewLogger() *Logger {
	c := newLoggerConfig()
	logger, _ := c.Build()
	defer logger.Sync()

	return &Logger{
		SugaredLogger: logger.Sugar(),
		dbLogging:     true,
	}
}

// NewFakeLogger initializes a fake Logger instance that is useful for testing purpose.
func NewFakeLogger() (*Logger, *bytes.Buffer, *bufio.Writer) {
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

// BeforeQuery is a hook before a go-pg's DB query.
func (l Logger) BeforeQuery(c context.Context, e *DBQueryEvent) (context.Context, error) {
	return c, nil
}

// AfterQuery is a hook after a go-pg's DB query.
func (l Logger) AfterQuery(c context.Context, e *DBQueryEvent) error {
	query, err := e.FormattedQuery()

	if !strings.Contains(query, dbQueryComment) && l.dbLogging {
		replacer := strings.NewReplacer("\n", "", ",\n", ", ", "\t", "")
		l.SugaredLogger.Infof("[SQL] %s in %s", replacer.Replace(query), time.Since(e.StartTime))
	}

	return err
}

// DBLogging can be used to check if DB logging is enabled or not.
func (l Logger) DBLogging() bool {
	return l.dbLogging
}

// SetDBLogging can be used to toggle the DB logging.
func (l *Logger) SetDBLogging(enabled bool) {
	l.dbLogging = enabled
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
