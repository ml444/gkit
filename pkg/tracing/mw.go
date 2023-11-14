package tracing

/*
import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/pkg/header"
	"github.com/ml444/gkit/transport"
)

// GRPCServerInterceptor returns a new server middleware for OpenTelemetry.
func GRPCServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	tracer := NewTracer(trace.SpanKindServer, opts...)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (reply interface{}, err error) {
		tr, ok := transport.FromContext(ctx)
		if !ok {
			tr = transport.GetTransportFromGrpcServer(ctx, info, transport.Metadata{})
			ctx = transport.ToContext(ctx, tr)
		}
		traceId := header.GetTraceId(ctx)
		if traceId == "" {
			var span trace.Span
			ctx, span = tracer.Start(ctx, tr.GetOperation(), tr.GetReqHeader())
			defer func() { tracer.End(ctx, span, reply, err) }()
			traceId = TraceID(ctx)
			log.Debugf("===> new trace_id: %s", traceId)
			SetServerSpan(ctx, span, req)
		}
		//spanCtx := trace.SpanContextFromContext(ctx)
		CacheTraceId.SetTraceIdWithoutKey(traceId)
		defer func() { CacheTraceId.SetTraceIdWithoutKey("") }()
		return handler(ctx, req)
	}
}

// GRPCClientInterceptor returns a new client middleware for OpenTelemetry.
func GRPCClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	tracer := NewTracer(trace.SpanKindClient, opts...)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		tr, ok := transport.FromContext(ctx)
		if !ok {
			tr = transport.GetTransportFromGrpcClient(ctx, method, cc, transport.Metadata{})
			ctx = transport.ToContext(ctx, tr)
		}
		traceId := header.GetTraceId(ctx)
		if traceId == "" {
			var span trace.Span
			ctx, span = tracer.Start(ctx, tr.GetOperation(), tr.GetReqHeader())
			defer func() { tracer.End(ctx, span, reply, err) }()
			traceId = TraceID(ctx)
			log.Debugf("===> new trace_id: %s", traceId)
			SetClientSpan(ctx, span, req)
		}
		CacheTraceId.SetTraceIdWithoutKey(traceId)
		defer func() { CacheTraceId.SetTraceIdWithoutKey("") }()
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
func HTTPMiddleware(opts ...Option) middleware.HttpMiddleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			ctx := request.Context()
			tr, ok := transport.FromContext(ctx)
			if !ok {
				tr = transport.GetTransportFromHTTP(request)
				ctx = transport.ToContext(ctx, tr)
			}

			tracer := NewTracer(trace.SpanKindServer, opts...)
			var span trace.Span
			ctx, span = tracer.Start(ctx, tr.GetOperation(), tr.GetReqHeader())
			defer func() { tracer.End(ctx, span, nil, nil) }()
			traceId := TraceID(ctx)
			writer.Header().Set(header.TraceIdKey, traceId)
			header.SetTraceId2Headers(request.Header, traceId)
			CacheTraceId.SetTraceIdWithoutKey(traceId)
			defer func() { CacheTraceId.SetTraceIdWithoutKey("") }()
			log.Debugf("===>Server TraceId: %s", CacheTraceId.GetTraceIdWithoutKey())

			request = request.WithContext(ctx)
			handler.ServeHTTP(writer, request)
		})
	}
}

*/
