package support_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/appist/appy/support"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func TestLogger(t *testing.T) {
	t.Run("should return a method called 'Desugar'", func(t *testing.T) {
		logger := support.NewLogger()
		_, ok := reflect.TypeOf(logger).MethodByName("Desugar")

		assert.Equal(t, ok, true)
	})

	t.Run("should print with color code when APP_ENV is not set", func(t *testing.T) {
		appEnv, exists := os.LookupEnv("APP_ENV")
		defer func() {
			if exists {
				os.Setenv("APP_ENV", appEnv)
			}
		}()
		os.Unsetenv("APP_ENV")

		spanCtx := trace.SpanContextFromContext(context.Background())
		spanCtx = spanCtx.WithSpanID(trace.SpanID([8]byte{1}))
		spanCtx = spanCtx.WithTraceID(trace.TraceID([16]byte{1}))
		ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)

		logger, buf, writer := support.NewTestLogger()
		logger.DebugContext(ctx, "test")
		logger.DebugfContext(ctx, "test %s", "foo")
		logger.ErrorContext(ctx, "test")
		logger.ErrorfContext(ctx, "test %s", "foo")
		logger.Info("test")
		logger.InfoContext(ctx, "test")
		logger.InfofContext(ctx, "test %s", "foo")
		logger.WarnContext(ctx, "test")
		logger.WarnfContext(ctx, "test %s", "foo")
		writer.Flush()

		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "\x1b[35mDEBUG\x1b[0m\ttest\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\x1b[35mDEBUG\x1b[0m\ttest foo\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\x1b[31mERROR\x1b[0m\ttest\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\x1b[31mERROR\x1b[0m\ttest foo\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\x1b[34mINFO\x1b[0m\ttest")
		assert.Contains(t, buf.String(), "\x1b[34mINFO\x1b[0m\ttest\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\x1b[34mINFO\x1b[0m\ttest foo\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\x1b[33mWARN\x1b[0m\ttest\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\x1b[33mWARN\x1b[0m\ttest foo\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
	})

	t.Run("should print with color code when APP_ENV is set to 'development'", func(t *testing.T) {
		appEnv, exists := os.LookupEnv("APP_ENV")
		defer func() {
			if exists {
				os.Setenv("APP_ENV", appEnv)
			}
		}()
		os.Setenv("APP_ENV", "development")

		spanCtx := trace.SpanContextFromContext(context.Background())
		spanCtx = spanCtx.WithSpanID(trace.SpanID([8]byte{1}))
		spanCtx = spanCtx.WithTraceID(trace.TraceID([16]byte{1}))
		ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)

		logger, buf, writer := support.NewTestLogger()
		logger.DebugContext(ctx, "test")
		logger.DebugfContext(ctx, "test %s", "foo")
		logger.ErrorContext(ctx, "test")
		logger.ErrorfContext(ctx, "test %s", "foo")
		logger.Info("test")
		logger.InfoContext(ctx, "test")
		logger.InfofContext(ctx, "test %s", "foo")
		logger.WarnContext(ctx, "test")
		logger.WarnfContext(ctx, "test %s", "foo")
		writer.Flush()

		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "\x1b[35mDEBUG\x1b[0m\ttest\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\x1b[35mDEBUG\x1b[0m\ttest foo\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\x1b[31mERROR\x1b[0m\ttest\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\x1b[31mERROR\x1b[0m\ttest foo\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\x1b[34mINFO\x1b[0m\ttest")
		assert.Contains(t, buf.String(), "\x1b[34mINFO\x1b[0m\ttest\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\x1b[34mINFO\x1b[0m\ttest foo\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\x1b[33mWARN\x1b[0m\ttest\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\x1b[33mWARN\x1b[0m\ttest foo\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
	})

	t.Run("should print without color code when APP_ENV is set to 'production'", func(t *testing.T) {
		appEnv, exists := os.LookupEnv("APP_ENV")
		defer func() {
			if exists {
				os.Setenv("APP_ENV", appEnv)
			}
		}()
		os.Setenv("APP_ENV", "production")

		spanCtx := trace.SpanContextFromContext(context.Background())
		spanCtx = spanCtx.WithSpanID(trace.SpanID([8]byte{1}))
		spanCtx = spanCtx.WithTraceID(trace.TraceID([16]byte{1}))
		ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)

		logger, buf, writer := support.NewTestLogger()
		logger.DebugContext(ctx, "test")
		logger.DebugfContext(ctx, "test %s", "foo")
		logger.ErrorContext(ctx, "test")
		logger.ErrorfContext(ctx, "test %s", "foo")
		logger.Info("test")
		logger.InfoContext(ctx, "test")
		logger.InfofContext(ctx, "test %s", "foo")
		logger.WarnContext(ctx, "test")
		logger.WarnfContext(ctx, "test %s", "foo")
		writer.Flush()

		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "\tdebug\ttest\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\tdebug\ttest foo\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\terror\ttest\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\terror\ttest foo\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\tinfo\ttest\n")
		assert.Contains(t, buf.String(), "\tinfo\ttest\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\tinfo\ttest foo\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\twarn\ttest\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
		assert.Contains(t, buf.String(), "\twarn\ttest foo\t{\"trace_id\": \"01000000000000000000000000000000\", \"span_id\": \"0100000000000000\"}\n")
	})
}
