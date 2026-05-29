package validate

import (
	"context"
	"errors"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
)

type IValidator interface {
	Validate() error
}

var invalidParamError = func(err error) error {
	var ex *errorx.Error
	if errors.As(err, &ex) {
		return errorx.FromError(err)
	}
	return errorx.CreateError(
		errorx.DefaultStatusCode,
		errorx.ErrCodeInvalidParamSys,
		err.Error(),
	).WithCause(err)
}

// Validator is a validator middleware.
func Validator() middleware.Middleware {
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			if v, ok := req.(IValidator); ok {
				if e := v.Validate(); e != nil {
					return rsp, invalidParamError(e)
				}
			}
			rsp, err = handler(ctx, req)
			if v, ok := rsp.(IValidator); ok {
				if e := v.Validate(); e != nil {
					return rsp, invalidParamError(e)
				}
			}
			return
		}
	}
}
