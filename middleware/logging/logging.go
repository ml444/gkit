package logging

import (
	"context"
	"time"

	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware"
)

func PrintInfo() middleware.Middleware {
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			log.Infof("req: %+v", req)
			startTime := time.Now()
			reply, err = handler(ctx, req)
			end := time.Since(startTime).Milliseconds()
			log.Infof("[%dms] rsp: %+v", end, reply)
			return
		}
	}
}
