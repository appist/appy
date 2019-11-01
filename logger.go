package appy

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	// Loggerer is the collection of method signatures for Logger struct.
	Loggerer interface {
		Debug(args ...interface{})
		Debugf(template string, args ...interface{})
		Debugw(msg string, keysAndValues ...interface{})
		Error(args ...interface{})
		Errorf(template string, args ...interface{})
		Errorw(msg string, keysAndValues ...interface{})
		Fatal(args ...interface{})
		Fatalf(template string, args ...interface{})
		Fatalw(msg string, keysAndValues ...interface{})
		Info(args ...interface{})
		Infof(template string, args ...interface{})
		Infow(msg string, keysAndValues ...interface{})
		Named(name string) *SugaredLogger
		DPanic(args ...interface{})
		DPanicf(template string, args ...interface{})
		DPanicw(msg string, keysAndValues ...interface{})
		Panic(args ...interface{})
		Panicf(template string, args ...interface{})
		Sync() error
		Warn(args ...interface{})
		Warnf(template string, args ...interface{})
		Warnw(msg string, keysAndValues ...interface{})
		With(args ...interface{}) *SugaredLogger
		BeforeQuery(c context.Context, e *DbQueryEvent) (context.Context, error)
		AfterQuery(c context.Context, e *DbQueryEvent) error
		Build() string
		DbLogging() bool
		SetDbLogging(enabled bool)
	}

	// SugaredLogger is a type alias to zap.SugaredLogger.
	SugaredLogger = zap.SugaredLogger

	// Logger provides the logging functionality.
	Logger struct {
		*SugaredLogger
		build     string
		dbLogging bool
	}
)

// NewLogger initializes Logger instance.
func NewLogger(build string) *Logger {
	c := newLoggerConfig(build)
	logger, _ := c.Build()
	defer logger.Sync()

	return &Logger{
		SugaredLogger: logger.Sugar(),
		build:         build,
		dbLogging:     true,
	}
}

// BeforeQuery is a hook before a db query.
func (l Logger) BeforeQuery(c context.Context, e *DbQueryEvent) (context.Context, error) {
	return c, nil
}

// AfterQuery is a hook after a db query.
func (l Logger) AfterQuery(c context.Context, e *DbQueryEvent) error {
	query, err := e.FormattedQuery()

	if !strings.Contains(query, "/* appy framework */") && l.dbLogging {
		replacer := strings.NewReplacer("\n", "", ",\n", ", ", "\t", "")
		l.SugaredLogger.Infof("[DB] %s in %s", replacer.Replace(query), time.Since(e.StartTime))
	}

	return err
}

// Build indicates if the logger is using debug or release config.
func (l Logger) Build() string {
	return l.build
}

// DbLogging can be used to check if DB logging is enabled or not.
func (l Logger) DbLogging() bool {
	return l.dbLogging
}

// SetDbLogging can be used to toggle the DB logging.
func (l *Logger) SetDbLogging(enabled bool) {
	l.dbLogging = enabled
}

func newLoggerConfig(build string) zap.Config {
	c := zap.NewDevelopmentConfig()
	c.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	c.EncoderConfig.CallerKey = ""
	c.EncoderConfig.EncodeTime = nil

	if build == ReleaseBuild {
		c = zap.NewProductionConfig()
		c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	return c
}
