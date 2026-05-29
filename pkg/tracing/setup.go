package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// DefaultOTLPEndpoint is the OTLP HTTP collector (Jaeger 1.35+, Grafana Tempo, OTel Collector).
const DefaultOTLPEndpoint = "localhost:4318"

// Config configures global OpenTelemetry for a gkit application.
type Config struct {
	// ServiceName is required and appears in trace backends.
	ServiceName string
	// OTLPEndpoint is host:port for OTLP/HTTP, e.g. "localhost:4318".
	// Empty uses OTEL_EXPORTER_OTLP_ENDPOINT (and related OTEL_* env vars).
	OTLPEndpoint string
	// OTLPInsecure uses http:// instead of https:// (typical for local collectors).
	OTLPInsecure bool
	// DisableExporter skips remote export (only in-process tracing / tests).
	DisableExporter bool
	// SampleRatio in [0,1]; default 1.0 (always sample root spans).
	SampleRatio float64
	// Propagator overrides W3C tracecontext + baggage (optional).
	Propagator propagation.TextMapPropagator
}

// Setup installs a global TracerProvider with an OTLP/HTTP exporter by default.
// Returns shutdown to flush spans. Jaeger is not used; point Jaeger 1.35+ at OTLP port 4318.
func Setup(cfg Config) (func(context.Context) error, error) {
	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("tracing: ServiceName is required")
	}
	ratio := cfg.SampleRatio
	if ratio <= 0 {
		ratio = 1.0
	}
	if ratio > 1 {
		ratio = 1.0
	}

	res := resource.NewSchemaless(
		semconv.ServiceNameKey.String(cfg.ServiceName),
		attribute.String("telemetry.sdk", "gkit/pkg/tracing"),
	)

	var tp *tracesdk.TracerProvider
	if cfg.DisableExporter {
		tp = tracesdk.NewTracerProvider(
			tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(ratio))),
			tracesdk.WithResource(res),
		)
	} else {
		exp, err := newOTLPExporter(context.Background(), cfg)
		if err != nil {
			return nil, fmt.Errorf("tracing: otlp exporter: %w", err)
		}
		tp = tracesdk.NewTracerProvider(
			tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(ratio))),
			tracesdk.WithBatcher(exp),
			tracesdk.WithResource(res),
		)
	}

	otel.SetTracerProvider(tp)
	if cfg.Propagator != nil {
		otel.SetTextMapPropagator(cfg.Propagator)
	}
	return tp.Shutdown, nil
}

func newOTLPExporter(ctx context.Context, cfg Config) (tracesdk.SpanExporter, error) {
	opts := []otlptracehttp.Option{}
	if cfg.OTLPEndpoint != "" {
		opts = append(opts, otlptracehttp.WithEndpoint(cfg.OTLPEndpoint))
	}
	if cfg.OTLPInsecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	return otlptracehttp.New(ctx, opts...)
}

// SyncTraceIDToCache copies the active OTel trace ID into the goroutine cache.
// Use only when integrating legacy loggers that read CacheTraceId (see README).
func SyncTraceIDToCache(ctx context.Context) {
	if CacheTraceId == nil {
		return
	}
	CacheTraceId.SetTraceIdWithoutKey(TraceID(ctx))
}

// ClearTraceIDCache removes the trace ID for the current goroutine.
func ClearTraceIDCache() {
	if CacheTraceId != nil {
		CacheTraceId.SetTraceIdWithoutKey("")
	}
}

// Provider returns the global tracer provider if configured.
func Provider() trace.TracerProvider {
	return otel.GetTracerProvider()
}
