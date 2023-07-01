package routine

import (
	"context"
	"fmt"
	"github.com/ml444/gkit/log"
	"runtime/debug"
	"strings"
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

func Go(ctx *context.Context, logic func(ctx *context.Context) error) {
	name := log.GetLoggerName()
	go func() {
		defer CatchPanic(func(err interface{}) {
			msg := fmt.Sprintf("%v %s: catch panic in go-routine, err %v", ctx, name, err)
			log.Error(msg)
		})
		err := logic(ctx)
		if err != nil {
			msg := fmt.Sprintf("%s: go-routine err %v", name, err)
			log.Error(msg)
		}
	}()
}
