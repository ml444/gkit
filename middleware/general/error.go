package general

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/pkg/errors"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
)

// WrapError: Unify errors into the Error format of errorx.
func WrapError() middleware.Middleware {
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			rsp, err = handler(ctx, req)
			if err != nil {
				return rsp, errorx.FromError(err)
			}
			return
		}
	}
}

func ServerErrorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	return resp, toStatusError(err)
}

func toStatusError(err error) error {
	if err == nil {
		return nil
	}
	cause := errors.Cause(err)
	pbErr, ok := cause.(*errorx.Error)
	if !ok {
		pbErr = errorx.CreateError(errorx.UnknownStatusCode, errorx.ErrCodeUnknown, cause.Error())
	}
	st := status.New(codes.Internal, cause.Error())
	st, e := st.WithDetails(pbErr)
	if e != nil {
		// make sure pbErr implements proto.Message interface
		return errorx.CreateErrorf(
			errorx.UnknownStatusCode,
			errorx.ErrCodeUnknown,
			"errorx: %s to %s", pbErr.String(), e.Error(),
		)
	}
	return st.Err()
}

func ClientErrorInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	err := invoker(ctx, method, req, reply, cc, opts...)
	if err == nil {
		return nil
	}
	cause := errors.Cause(err)
	st, ok := status.FromError(cause)
	if ok {
		details := st.Details()
		for _, detail := range details {
			if pbErr, ok := detail.(*errorx.Error); ok {
				return errorx.CreateError(pbErr.StatusCode, pbErr.ErrorCode, pbErr.Message)
			}
		}
	}
	return err
}
