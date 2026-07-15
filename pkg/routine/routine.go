package routine

import (
	"context"
	"errors"
	"runtime/debug"
	"strings"

	"github.com/ml444/gkit/log"
)

func CatchPanic(cb func(err interface{})) {
	if err := recover(); err != nil {
		log.Errorf("PROCESS PANIC: err %s", err)
		st := debug.Stack()
		if len(st) > 0 {
			log.Errorf("dump stack (%s):", err)
			lines := strings.Split(string(st), "\n")
			for _, line := range lines {
				log.Error("  ", line)
			}
		} else {
			log.Errorf("stack is empty (%s)", err)
		}
		if cb != nil {
			cb(err)
		}
	}
}

// Go runs fn in a new goroutine with panic recovery.
// If ctx is already cancelled or timed out, fn is not started.
// fn should respect ctx cancellation during blocking I/O.
func Go(ctx context.Context, fn func(context.Context) error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return
	}
	name := log.GetLoggerName()
	go func() {
		defer CatchPanic(func(err interface{}) {
			log.Errorf("%s: catch panic in goroutine, err %v", name, err)
		})
		if err := fn(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Errorf("%s: goroutine err %v", name, err)
		}
	}()
}
