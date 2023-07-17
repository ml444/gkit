package grpcmw

import (
	"context"
	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware"
	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a new unary server interceptor for panic recovery.
func UnaryServerInterceptor(handlers ...middleware.HandlerFuncContext) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ any, err error) {
		defer func() {
			if r := recover(); r != nil {
				for _, h := range handlers {
					err = h(ctx, r)
					if err != nil {
						log.Errorf("panic recovery failed: %v", err)
					}
					return
				}
			}
		}()

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor for panic recovery.
func StreamServerInterceptor(handlers ...middleware.HandlerFuncContext) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				for _, h := range handlers {
					err = h(stream.Context(), r)
					if err != nil {
						log.Errorf("panic recovery failed: %v", err)
						return
					}
				}
			}
		}()

		return handler(srv, stream)
	}
}
