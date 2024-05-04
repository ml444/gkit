package recovery

import (
	"context"
	"fmt"
	"runtime"

	"google.golang.org/grpc"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware"
)

type (
	RecoverKey struct{}
	RequestKey struct{}
)

func Recovery(fns ...middleware.LurkerFunc) middleware.Middleware {
	if len(fns) == 0 {
		fns = append(fns, func(ctx context.Context, buf interface{}) error {
			req := ctx.Value(RequestKey{})
			r := ctx.Value(RecoverKey{})
			log.Errorf("%v: %+v\n%s\n", r, req, buf)
			return errorx.InternalServer(fmt.Sprintf("%v", r))
		})
	}
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 64<<10) //nolint:gomnd
					n := runtime.Stack(buf, false)
					buf = buf[:n]
					ctx = context.WithValue(ctx, RecoverKey{}, r)
					ctx = context.WithValue(ctx, RequestKey{}, req)
					err = middleware.LurkerChain(ctx, r)
				}
			}()
			return handler(ctx, req)
		}
	}
}

// UnaryServerInterceptor returns a new unary server interceptor for panic recovery.
func UnaryServerInterceptor(fns ...middleware.LurkerFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ any, err error) {
		defer func() {
			if r := recover(); r != nil {
				for _, h := range fns {
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
func StreamServerInterceptor(fns ...middleware.LurkerFunc) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				for _, h := range fns {
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
