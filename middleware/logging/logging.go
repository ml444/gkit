package logging

import (
	"context"
	"time"

	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/pkg/header"
	"github.com/ml444/gkit/transport"
)

type (
	Took  struct{}
	Reply struct{}
)

func defaultLogging() middleware.LurkerFunc {
	return func(ctx context.Context, req any) error {
		took := ctx.Value(Took{})
		path := ""
		if tr, ok := transport.FromContext(ctx); ok {
			path = tr.Path()
		}
		ti := header.TraceInfoFromContext(ctx)
		trace := ti.TraceID
		if trace == "" {
			trace = header.CorrelationID(ctx)
		}
		log.Infof("trace=%s span=%s path=%s took=%vms req=%+v", trace, ti.SpanID, path, took, req)
		return nil
	}
}

// LogRequest logs request latency and payload after handler completes.
func LogRequest(fns ...middleware.LurkerFunc) middleware.Middleware {
	if len(fns) == 0 {
		fns = append(fns, defaultLogging())
	}
	return func(handler middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (rsp interface{}, err error) {
			startTime := time.Now()
			rsp, err = handler(ctx, req)
			ctx = context.WithValue(ctx, Took{}, time.Since(startTime).Milliseconds())
			ctx = context.WithValue(ctx, Reply{}, rsp)
			if err != nil {
				log.Errorf("request failed: %v", err)
			}
			middleware.ForceLurkerChain(ctx, req, fns...)
			return
		}
	}
}
