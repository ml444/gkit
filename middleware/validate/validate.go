package validate

import (
	"context"
	"github.com/ml444/gkit/errors"
	"github.com/ml444/gkit/middleware"
)

type validator interface {
	Validate() error
}

// Validator is a validator middleware.
func Validator() middleware.Middleware {
	return func(handler middleware.Endpoint) middleware.Endpoint {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if v, ok := req.(validator); ok {
				if err := v.Validate(); err != nil {
					return nil, errors.CreateError(errors.DefaultStatusCode, errors.ErrCodeInvalidParamSys, err.Error()).WithCause(err)
				}
			}
			return handler(ctx, req)
		}
	}
}
