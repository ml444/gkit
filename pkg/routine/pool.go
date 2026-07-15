package routine

import (
	"context"
	"errors"
	"fmt"

	"github.com/ml444/gkit/log"
)

// Pool limits the number of concurrent background goroutines.
type Pool struct {
	sem chan struct{}
}

// NewPool creates a pool that allows at most maxWorkers concurrent tasks.
// Returns an error when maxWorkers is less than 1.
func NewPool(maxWorkers int) (*Pool, error) {
	if maxWorkers < 1 {
		return nil, fmt.Errorf("routine: maxWorkers must be >= 1, got %d", maxWorkers)
	}
	return &Pool{sem: make(chan struct{}, maxWorkers)}, nil
}

// Go runs fn in a background goroutine when a pool slot is available.
// If the pool is full, it blocks until a slot opens or ctx is cancelled.
// If ctx is already cancelled, fn is not started and ctx.Err() is returned.
func (p *Pool) Go(ctx context.Context, fn func(context.Context) error) error {
	if p == nil {
		return fmt.Errorf("routine: nil pool")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	select {
	case p.sem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	name := log.GetLoggerName()
	go func() {
		defer func() { <-p.sem }()
		defer CatchPanic(func(err interface{}) {
			log.Errorf("%s: catch panic in goroutine, err %v", name, err)
		})
		if err := fn(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Errorf("%s: goroutine err %v", name, err)
		}
	}()
	return nil
}
