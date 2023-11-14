package general

import (
	"context"
	"reflect"

	"github.com/ml444/gkit/middleware"
)

func ReplaceEmptyResponse(data interface{}) middleware.Middleware {
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			rsp, err = handler(ctx, req)
			if err == nil && (rsp == nil || reflect.ValueOf(rsp).Elem().IsZero()) {
				return data, nil
			}
			return
		}
	}
}
