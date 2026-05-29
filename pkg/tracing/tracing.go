package tracing

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Option is tracing option.
type Option func(*options)

type options struct {
	tracerName     string
	tracerProvider trace.TracerProvider
	propagator     propagation.TextMapPropagator
}

// WithPropagator with tracer propagator.
func WithPropagator(propagator propagation.TextMapPropagator) Option {
	return func(opts *options) {
		opts.propagator = propagator
	}
}

// WithTracerProvider with tracer provider.
// By default, it uses the global provider that is set by otel.SetTracerProvider(provider).
func WithTracerProvider(provider trace.TracerProvider) Option {
	return func(opts *options) {
		opts.tracerProvider = provider
	}
}

// WithTracerName with tracer name
func WithTracerName(tracerName string) Option {
	return func(opts *options) {
		opts.tracerName = tracerName
	}
}

// TraceID returns a trace_id valuer.
func TraceID(ctx context.Context) string {
	if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
		return span.TraceID().String()
	}
	return ""
}

// SpanID returns a span_id valuer.
func SpanID(ctx context.Context) string {
	if span := trace.SpanContextFromContext(ctx); span.HasSpanID() {
		return span.SpanID().String()
	}
	return ""
}

// InitTracer sets the global TracerProvider with OTLP export.
//
// Deprecated: use Setup(Config{...}) which returns shutdown.
// The url parameter is treated as OTLP endpoint host:port (not legacy Jaeger /api/traces URL).
// If url is empty, OTEL_EXPORTER_OTLP_ENDPOINT or DefaultOTLPEndpoint is used.
func InitTracer(url string, serviceName string) error {
	cfg := Config{
		ServiceName:  serviceName,
		SampleRatio:  1.0,
		OTLPInsecure: true,
	}
	if url != "" {
		cfg.OTLPEndpoint = url
	} else {
		cfg.OTLPEndpoint = DefaultOTLPEndpoint
	}
	_, err := Setup(cfg)
	return err
}
