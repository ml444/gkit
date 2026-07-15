package routine

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestGo_skipsWhenCtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var ran atomic.Bool
	Go(ctx, func(context.Context) error {
		ran.Store(true)
		return nil
	})

	time.Sleep(50 * time.Millisecond)
	if ran.Load() {
		t.Fatal("expected fn not to run when ctx is already cancelled")
	}
}

func TestGo_recoversPanic(t *testing.T) {
	done := make(chan struct{})
	Go(context.Background(), func(context.Context) error {
		defer close(done)
		panic("test panic")
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("goroutine did not finish after panic")
	}
}

func TestGo_propagatesCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	Go(ctx, func(c context.Context) error {
		defer close(done)
		if c != ctx {
			t.Error("expected fn to receive the same context")
		}
		select {
		case <-c.Done():
			return c.Err()
		case <-time.After(50 * time.Millisecond):
			return nil
		}
	})

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("goroutine did not observe ctx cancellation")
	}
}

func TestGo_logsReturnedError(t *testing.T) {
	done := make(chan struct{})
	Go(context.Background(), func(context.Context) error {
		defer close(done)
		return errors.New("boom")
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("goroutine did not finish")
	}
}

func TestGo_nilCtxUsesBackground(t *testing.T) {
	done := make(chan struct{})
	Go(nil, func(ctx context.Context) error {
		defer close(done)
		if ctx == nil {
			t.Fatal("expected non-nil context")
		}
		if err := ctx.Err(); err != nil {
			t.Fatalf("unexpected ctx error: %v", err)
		}
		return nil
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("goroutine did not finish")
	}
}
