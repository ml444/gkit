package httpmw

import (
	"context"
	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
)

type Validater interface {
	Validate() error
}

// Validator is a validator middleware.
func Validator() middleware.BeforeHandler {
	return func(ctx context.Context, req interface{}) (context.Context, interface{}, error) {
		if v, ok := req.(Validater); ok {
			if err := v.Validate(); err != nil {
				return ctx, nil, errorx.CreateError(
					errorx.DefaultStatusCode,
					errorx.ErrCodeInvalidParamSys,
					err.Error(),
				).WithCause(err)
			}
		}
		return ctx, req, nil
	}
}
