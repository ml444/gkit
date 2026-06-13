package recovery

import (
	"context"
	"fmt"
	"net/http"
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

func defaultRecoveryHandler() middleware.LurkerFunc {
	return func(ctx context.Context, p interface{}) error {
		req := ctx.Value(RequestKey{})
		r := ctx.Value(RecoverKey{})
		log.Errorf("%v: %+v\n%s\n", r, req, p)
		return errorx.InternalServer(fmt.Sprintf("%v", r))
	}
}

func Recovery(fns ...middleware.LurkerFunc) middleware.Middleware {
	if len(fns) == 0 {
		fns = []middleware.LurkerFunc{defaultRecoveryHandler()}
	}
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 64<<10) //nolint:gomnd
					n := runtime.Stack(buf, false)
					buf = buf[:n]
					pctx := context.WithValue(ctx, RecoverKey{}, r)
					pctx = context.WithValue(pctx, RequestKey{}, req)
					err = middleware.LurkerChain(pctx, buf, fns...)
				}
			}()
			return handler(ctx, req)
		}
	}
}

// HTTPMiddleware returns HTTP middleware that recovers from panics in HTTP
// handlers (and HTTP-only middleware) and responds with 500.
func HTTPMiddleware(fns ...middleware.LurkerFunc) middleware.HttpMiddleware {
	if len(fns) == 0 {
		fns = []middleware.LurkerFunc{defaultRecoveryHandler()}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					buf := make([]byte, 64<<10) //nolint:gomnd
					n := runtime.Stack(buf, false)
					buf = buf[:n]
					pctx := context.WithValue(r.Context(), RecoverKey{}, rec)
					pctx = context.WithValue(pctx, RequestKey{}, r.URL.Path)
					_ = middleware.LurkerChain(pctx, buf, fns...)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// UnaryServerInterceptor returns a new unary server interceptor for panic recovery.
func UnaryServerInterceptor(fns ...middleware.LurkerFunc) grpc.UnaryServerInterceptor {
	if len(fns) == 0 {
		fns = []middleware.LurkerFunc{defaultRecoveryHandler()}
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ any, err error) {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64<<10) //nolint:gomnd
				n := runtime.Stack(buf, false)
				buf = buf[:n]
				pctx := context.WithValue(ctx, RecoverKey{}, r)
				pctx = context.WithValue(pctx, RequestKey{}, req)
				err = middleware.LurkerChain(pctx, buf, fns...)
			}
		}()

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor for panic recovery.
func StreamServerInterceptor(fns ...middleware.LurkerFunc) grpc.StreamServerInterceptor {
	if len(fns) == 0 {
		fns = []middleware.LurkerFunc{defaultRecoveryHandler()}
	}
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 64<<10) //nolint:gomnd
				n := runtime.Stack(buf, false)
				buf = buf[:n]
				ctx := stream.Context()
				pctx := context.WithValue(ctx, RecoverKey{}, r)
				pctx = context.WithValue(pctx, RequestKey{}, info.FullMethod)
				err = middleware.LurkerChain(pctx, buf, fns...)
			}
		}()

		return handler(srv, stream)
	}
}
