package support

import (
	"bufio"
	"bytes"
	"context"
	"strings"
	"time"

	"github.com/go-pg/pg/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	// SugaredLogger is a type alias to zap.SugaredLogger.
	SugaredLogger = zap.SugaredLogger

	// Logger provides the logging functionality.
	Logger struct {
		*SugaredLogger
		dbLogging bool
	}
)

const (
	// DbQueryComment is used to identify which DB query is trigerred from within appy framework.
	DbQueryComment = "/* appy framework */"
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

// BeforeQuery is a hook before a go-pg's DB query.
func (l Logger) BeforeQuery(c context.Context, e *pg.QueryEvent) (context.Context, error) {
	return c, nil
}

// AfterQuery is a hook after a go-pg's DB query.
func (l Logger) AfterQuery(c context.Context, e *pg.QueryEvent) error {
	query, err := e.FormattedQuery()

	if !strings.Contains(query, DbQueryComment) && l.dbLogging {
		replacer := strings.NewReplacer("\n", "", ",\n", ", ", "\t", "")
		l.SugaredLogger.Infof("[DB] %s in %s", replacer.Replace(query), time.Since(e.StartTime))
	}

	return err
}

// DbLogging can be used to check if DB logging is enabled or not.
func (l Logger) DbLogging() bool {
	return l.dbLogging
}

// SetDbLogging can be used to toggle the DB logging.
func (l *Logger) SetDbLogging(enabled bool) {
	l.dbLogging = enabled
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
