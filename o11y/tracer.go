package o11y

import (
	"context"
	"net"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type (

	// Span wraps up trace.Span which is the individual component of a trace. It represents a single
	// named and timed operation of a workflow that is traced. A Tracer is used to create a Span and
	// it is then up to the operation the Span represents to properly end the Span when the operation
	// itself ends.
	Span struct {
		trace.Span
	}

	// Tracer wraps up trace.Tracer which is the creator of Spans.
	Tracer struct {
		trace.Tracer
	}

	// Tracer wraps up trace.TracerOption and applies an option to a TracerConfig.
	TracerOption struct {
		trace.TracerOption
	}

	// TracerProvider provides access to instrumentation Tracers.
	TracerProvider struct {
		*sdktrace.TracerProvider
	}

	// TracerProviderConfig indicates how the OpenTelemetry tracer provider should be initialised.
	TracerProviderConfig struct {
		// ServiceName indicates the service name to trace.
		ServiceName string

		// CollectorAddress indicates the collector's address.
		CollectorAddress string

		// CollectorConnectTimeout indicate the duration to timeout when connecting to the collector.
		// By default, it is 5 * time.Second.
		CollectorConnectTimeout time.Duration
	}
)

// NewTracerProvider initializes the provider for OpenTelemetry traces.
func NewTracerProvider(c *TracerProviderConfig) (*TracerProvider, func() error, error) {
	ctx := context.Background()
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(c.ServiceName),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	defaultTracerProviderConfig(c)

	// Initialise the tracer provider which will use the exporter to push traces to the collector.
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(c.CollectorAddress),
		otlptracegrpc.WithDialOption(
			grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
				return net.DialTimeout("tcp", addr, c.CollectorConnectTimeout)
			}),
		),
	)

	if err != nil {
		return nil, nil, err
	}

	batchSpanProcessor := sdktrace.NewBatchSpanProcessor(exporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(batchSpanProcessor),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return &TracerProvider{tracerProvider}, func() error {
		if err := exporter.Shutdown(ctx); err != nil {
			return err
		}

		if err := tracerProvider.Shutdown(ctx); err != nil {
			return err
		}

		return nil
	}, nil
}

// NewTracer initialises a tracer.
func NewTracer(instrumentationName string, opts ...trace.TracerOption) Tracer {
	provider := otel.GetTracerProvider()

	return Tracer{
		provider.Tracer(instrumentationName, opts...),
	}
}

// SpanFromContext returns the current Span from ctx.
//
// If no Span is currently set in ctx an implementation of a Span that performs no operations is
// returned.
func SpanFromContext(ctx context.Context) trace.Span {
	return Span{
		trace.SpanFromContext(ctx),
	}
}

func GetSpanIDFromContext(ctx context.Context) string {
	return SpanFromContext(ctx).SpanContext().SpanID().String()
}

func GetTraceIDFromContext(ctx context.Context) string {
	return SpanFromContext(ctx).SpanContext().TraceID().String()
}

func defaultTracerProviderConfig(c *TracerProviderConfig) {
	if c.CollectorConnectTimeout == 0 {
		c.CollectorConnectTimeout = 5 * time.Second
	}
}
