package support

import (
	"bufio"
	"bytes"
	"context"
	"os"

	"github.com/appist/appy/o11y"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	spanId  = "span_id"
	traceId = "trace_id"
)

// Logger provides the logging functionality.
type Logger struct {
	*zap.SugaredLogger
}

// NewLogger initialises a logger instance.
func NewLogger() *Logger {
	c := newLoggerConfig()
	logger, _ := c.Build()

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}
}

// DebugContext uses fmt.Sprint to construct and log a message with the `trace_id` found in the context.
func (logger *Logger) DebugContext(ctx context.Context, args ...interface{}) {
	logger.With(traceId, o11y.GetTraceIDFromContext(ctx)).With(spanId, o11y.GetSpanIDFromContext(ctx)).Debug(args...)
}

// DebugfContext uses fmt.Sprintf to log a templated message with the `trace_id` found in the context.
func (logger *Logger) DebugfContext(ctx context.Context, template string, args ...interface{}) {
	logger.With(traceId, o11y.GetTraceIDFromContext(ctx)).With(spanId, o11y.GetSpanIDFromContext(ctx)).Debugf(template, args...)
}

// ErrorContext uses fmt.Sprint to construct and log a message with the `trace_id` found in the context.
func (logger *Logger) ErrorContext(ctx context.Context, args ...interface{}) {
	logger.With(traceId, o11y.GetTraceIDFromContext(ctx)).With(spanId, o11y.GetSpanIDFromContext(ctx)).Error(args...)
}

// ErrorfContext uses fmt.Sprintf to log a templated message with the `trace_id` found in the context.
func (logger *Logger) ErrorfContext(ctx context.Context, template string, args ...interface{}) {
	logger.With(traceId, o11y.GetTraceIDFromContext(ctx)).With(spanId, o11y.GetSpanIDFromContext(ctx)).Errorf(template, args...)
}

// InfoContext uses fmt.Sprint to construct and log a message with the `trace_id` found in the context.
func (logger *Logger) InfoContext(ctx context.Context, args ...interface{}) {
	logger.With(traceId, o11y.GetTraceIDFromContext(ctx)).With(spanId, o11y.GetSpanIDFromContext(ctx)).Info(args...)
}

// InfofContext uses fmt.Sprintf to log a templated message with the `trace_id` found in the context.
func (logger *Logger) InfofContext(ctx context.Context, template string, args ...interface{}) {
	logger.With(traceId, o11y.GetTraceIDFromContext(ctx)).With(spanId, o11y.GetSpanIDFromContext(ctx)).Infof(template, args...)
}

// WarnContext uses fmt.Sprint to construct and log a message with the `trace_id` found in the context.
func (logger *Logger) WarnContext(ctx context.Context, args ...interface{}) {
	logger.With(traceId, o11y.GetTraceIDFromContext(ctx)).With(spanId, o11y.GetSpanIDFromContext(ctx)).Warn(args...)
}

// WarnfContext uses fmt.Sprintf to log a templated message with the `trace_id` found in the context.
func (logger *Logger) WarnfContext(ctx context.Context, template string, args ...interface{}) {
	logger.With(traceId, o11y.GetTraceIDFromContext(ctx)).With(spanId, o11y.GetSpanIDFromContext(ctx)).Warnf(template, args...)
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
	c.EncoderConfig.TimeKey = ""
	c.EncoderConfig.CallerKey = ""

	if os.Getenv("APP_ENV") != "" && os.Getenv("APP_ENV") != "development" {
		c = zap.NewProductionConfig()
	}

	return c
}
