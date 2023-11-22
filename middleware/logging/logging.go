package logging

import (
	"context"
	"time"

	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware"
)

func PrintInfo(printRspBody bool) middleware.Middleware {
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			log.Infof("req: %+v \n", req)
			startTime := time.Now()
			rsp, err = handler(ctx, req)
			end := time.Since(startTime).Milliseconds()
			if printRspBody {
				log.Infof("[%dms] rsp: %+v \n", end, rsp)
			} else {
				log.Infof("[%dms] \n", end)
			}
			return
		}
	}
}
