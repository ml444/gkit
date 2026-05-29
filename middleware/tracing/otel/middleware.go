// Package otel provides OpenTelemetry middleware as an optional submodule.
//
// The main gkit module uses lightweight trace ID propagation in middleware/tracing.
// Import this package when you need full OTel spans:
//
//	import "github.com/ml444/gkit/middleware/tracing/otel"
//
// Requires github.com/ml444/gkit/pkg/tracing (OpenTelemetry SDK) as a dependency.
package otel

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/pkg/header"
	gkitotel "github.com/ml444/gkit/pkg/tracing"
	"github.com/ml444/gkit/transport"
	"github.com/ml444/gkit/transport/httpx"
)

type mdCarrier transport.MD

func (c mdCarrier) Get(key string) string { return transport.MD(c).GetFirst(key) }
func (c mdCarrier) Set(key, value string) { transport.MD(c).Set(key, value) }
func (c mdCarrier) Keys() []string       { return transport.MD(c).Keys() }

// Server returns service middleware with OTel server spans.
func Server(opts ...gkitotel.Option) middleware.Middleware {
	tracer := gkitotel.NewTracer(trace.SpanKindServer, opts...)
	return func(next middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			tr, ok := transport.FromContext(ctx)
			if !ok {
				return next(ctx, req)
			}
			var carrier propagation.TextMapCarrier = mdCarrier{}
			if tr.In() != nil {
				carrier = mdCarrier(tr.In())
			}
			op := tr.Path()
			ctx, span := tracer.Start(ctx, op, carrier)
			defer func() { tracer.End(ctx, span, rsp, err) }()
			if id := gkitotel.TraceID(ctx); id != "" {
				ctx = header.WithTraceID(ctx, id)
				header.SetOutgoing(ctx, header.TraceIDKey, id)
				gkitotel.SyncTraceIDToCache(ctx)
			}
			defer gkitotel.ClearTraceIDCache()
			return next(ctx, req)
		}
	}
}

// HTTPMiddleware creates OTel spans for HTTP requests.
func HTTPMiddleware(opts ...gkitotel.Option) middleware.HttpMiddleware {
	tracer := gkitotel.NewTracer(trace.SpanKindServer, opts...)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tr, ok := transport.FromContext(r.Context())
			if !ok {
				tr = httpx.ClientTransport(r)
				r = r.WithContext(transport.ToContext(r.Context(), tr))
			}
			var carrier propagation.TextMapCarrier = mdCarrier{}
			if tr.In() != nil {
				carrier = mdCarrier(tr.In())
			}
			ctx, span := tracer.Start(r.Context(), tr.Path(), carrier)
			defer func() { tracer.End(ctx, span, nil, nil) }()
			if id := gkitotel.TraceID(ctx); id != "" {
				header.SetTraceID(w.Header(), id)
				ctx = header.WithTraceID(ctx, id)
				gkitotel.SyncTraceIDToCache(ctx)
			}
			defer gkitotel.ClearTraceIDCache()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UnaryServerInterceptor returns a gRPC unary interceptor with OTel spans.
func UnaryServerInterceptor(opts ...gkitotel.Option) grpc.UnaryServerInterceptor {
	tracer := gkitotel.NewTracer(trace.SpanKindServer, opts...)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		tr, ok := transport.FromContext(ctx)
		if !ok {
			return handler(ctx, req)
		}
		var carrier propagation.TextMapCarrier = mdCarrier{}
		if tr.In() != nil {
			carrier = mdCarrier(tr.In())
		}
		ctx, span := tracer.Start(ctx, info.FullMethod, carrier)
		var reply any
		var err error
		defer func() { tracer.End(ctx, span, reply, err) }()
		reply, err = handler(ctx, req)
		return reply, err
	}
}
