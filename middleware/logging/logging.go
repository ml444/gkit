package logging

import (
	"context"
	"time"

	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware"
)

type (
	Took  struct{}
	Reply struct{}
)

func defalutLogging() middleware.LurkerFunc {
	return func(ctx context.Context, req any) error {
		end := ctx.Value(Took{})
		log.Infof("took: [%dms] req: %+v \n", end, req)
		return nil
	}
}

func LogRequest(fns ...middleware.LurkerFunc) middleware.Middleware {
	if len(fns) == 0 {
		fns = append(fns, defalutLogging())
	}
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			startTime := time.Now()
			rsp, err = handler(ctx, req)
			ctx = context.WithValue(ctx, Took{}, time.Since(startTime).Milliseconds())
			ctx = context.WithValue(ctx, Reply{}, rsp)
			middleware.ForceLurkerChain(ctx, req, fns...)
			return
		}
	}
}
