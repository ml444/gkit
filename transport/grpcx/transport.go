package grpcx

import (
	"context"

	"google.golang.org/grpc"
	grpcmd "google.golang.org/grpc/metadata"

	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/pkg/header"
	"github.com/ml444/gkit/transport"
)

var _ transport.ITransport = (*Transport)(nil)

type Transport struct {
	transport.BaseTransport
}

func (tr *Transport) GetKind() transport.Kind {
	return transport.KindGRPC
}

/*
>>>>>>>>>>>> server interceptor <<<<<<<<<<<<<
*/

func (s *Server) unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, _ := grpcmd.FromIncomingContext(ctx)
		outHeader := grpcmd.MD{}
		tr := &Transport{
			BaseTransport: transport.BaseTransport{
				Endpoint:  s.endpoint,
				Operation: info.FullMethod,
				InHeader:  header.Header(md),
				OutHeader: header.Header(outHeader),
			},
		}
		ctx = transport.ToContext(ctx, tr)
		if s.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, s.timeout)
			defer cancel()
		}
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			return handler(ctx, req)
		}
		if len(s.middlewares) > 0 {
			h = middleware.Chain(s.middlewares...)(h)
		}
		reply, err := h(ctx, req)
		if len(outHeader) > 0 {
			_ = grpc.SetHeader(ctx, outHeader)
		}
		return reply, err
	}
}

// wrappedStream is rewrite grpc stream's context
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func NewWrappedStream(ctx context.Context, stream grpc.ServerStream) grpc.ServerStream {
	return &wrappedStream{
		ServerStream: stream,
		ctx:          ctx,
	}
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// streamServerInterceptor is a gRPC stream server interceptor
func (s *Server) streamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		md, _ := grpcmd.FromIncomingContext(ctx)
		outHeader := grpcmd.MD{}
		ctx = transport.ToContext(ctx, &Transport{
			BaseTransport: transport.BaseTransport{
				Endpoint:  s.endpoint,
				Operation: info.FullMethod,
				InHeader:  header.Header(md),
				OutHeader: header.Header(outHeader),
			},
		})

		ws := NewWrappedStream(ctx, ss)

		err := handler(srv, ws)
		if len(outHeader) > 0 {
			_ = grpc.SetHeader(ctx, outHeader)
		}
		return err
	}
}

/*
>>>>>>>>>>>>> Get transport <<<<<<<<<<<<<<<<
*/

func GetTransportFromGrpcClient(ctx context.Context, method string, cc *grpc.ClientConn, header header.IHeader) transport.ITransport {
	return &Transport{
		BaseTransport: transport.BaseTransport{
			Endpoint:  cc.Target(),
			Operation: method,
			InHeader:  header,
			OutHeader: nil,
		},
	}
}
