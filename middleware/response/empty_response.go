package response

import (
	"context"
	"reflect"

	"github.com/ml444/gkit/middleware"
)

// ReplaceEmptyResponse In some scenarios, when the response is empty,
// it needs to be modified into a unified format.
// For example: `{"err_code": 123000, "message": "success"}`
func isEmptyResponse(rsp interface{}) bool {
	if rsp == nil {
		return true
	}
	v := reflect.ValueOf(rsp)
	if v.Kind() != reflect.Ptr {
		return false
	}
	if v.IsNil() {
		return true
	}
	return v.Elem().IsZero()
}

func ReplaceEmptyResponse(data interface{}) middleware.Middleware {
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			rsp, err = handler(ctx, req)
			if err == nil && isEmptyResponse(rsp) {
				return data, nil
			}
			return
		}
	}
}
