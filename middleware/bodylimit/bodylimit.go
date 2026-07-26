package bodylimit

import (
	"context"
	"net/http"
	"reflect"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
)

var ErrBodyTooLarge = errorx.CreateError(413, 41301, "BODYLIMIT: request body too large")

func HTTPMiddleware(maxBytes int64) middleware.HttpMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// Server rejects requests whose serialized size exceeds maxBytes (approximate).
func Server(maxBytes int64) middleware.Middleware {
	return func(next middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if maxBytes > 0 && approxSize(req) > maxBytes {
				return nil, ErrBodyTooLarge
			}
			return next(ctx, req)
		}
	}
}

func approxSize(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case []byte:
		return int64(len(x))
	case string:
		return int64(len(x))
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr && !rv.IsNil() {
			if m, ok := v.(interface{ ProtoSize() int }); ok {
				return int64(m.ProtoSize())
			}
		}
	}
	return 0
}
