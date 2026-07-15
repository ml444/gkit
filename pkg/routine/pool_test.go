package routine

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewPool_invalidMaxWorkers(t *testing.T) {
	_, err := NewPool(0)
	if err == nil {
		t.Fatal("expected error for maxWorkers=0")
	}
	_, err = NewPool(-1)
	if err == nil {
		t.Fatal("expected error for negative maxWorkers")
	}
}

func TestPool_nilPool(t *testing.T) {
	var p *Pool
	err := p.Go(context.Background(), func(context.Context) error { return nil })
	if err == nil {
		t.Fatal("expected error for nil pool")
	}
}

func TestPool_limitsConcurrency(t *testing.T) {
	p, err := NewPool(2)
	if err != nil {
		t.Fatal(err)
	}

	var running atomic.Int32
	var maxSeen atomic.Int32
	started := make(chan struct{}, 3)
	release := make(chan struct{})

	runTask := func() error {
		return p.Go(context.Background(), func(context.Context) error {
			cur := running.Add(1)
			for {
				prev := maxSeen.Load()
				if cur <= prev || maxSeen.CompareAndSwap(prev, cur) {
					break
				}
			}
			started <- struct{}{}
			<-release
			running.Add(-1)
			return nil
		})
	}

	if err := runTask(); err != nil {
		t.Fatal(err)
	}
	if err := runTask(); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatalf("task %d did not start", i+1)
		}
	}

	if max := maxSeen.Load(); max > 2 {
		t.Fatalf("max concurrent workers = %d, want <= 2", max)
	}

	thirdDone := make(chan error, 1)
	go func() {
		thirdDone <- runTask()
	}()

	select {
	case <-started:
		t.Fatal("third task should not start before release")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("third task did not start after release")
	}

	select {
	case err := <-thirdDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("third Go did not return")
	}
}

func TestPool_blocksUntilSlot(t *testing.T) {
	p, err := NewPool(1)
	if err != nil {
		t.Fatal(err)
	}

	blockerStarted := make(chan struct{})
	release := make(chan struct{})
	if err := p.Go(context.Background(), func(context.Context) error {
		close(blockerStarted)
		<-release
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	select {
	case <-blockerStarted:
	case <-time.After(time.Second):
		t.Fatal("blocker did not start")
	}

	var secondRan atomic.Bool
	secondStarted := make(chan struct{})
	secondDone := make(chan error, 1)
	go func() {
		secondDone <- p.Go(context.Background(), func(context.Context) error {
			close(secondStarted)
			secondRan.Store(true)
			return nil
		})
	}()

	select {
	case <-secondStarted:
		t.Fatal("second task should block until slot is released")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)

	select {
	case <-secondStarted:
	case <-time.After(time.Second):
		t.Fatal("second task did not start after release")
	}

	select {
	case err := <-secondDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("second Go did not return")
	}
	if !secondRan.Load() {
		t.Fatal("expected second task to run")
	}
}

func TestPool_ctxCancelWhileWaiting(t *testing.T) {
	p, err := NewPool(1)
	if err != nil {
		t.Fatal(err)
	}

	release := make(chan struct{})
	blockerStarted := make(chan struct{})
	if err := p.Go(context.Background(), func(context.Context) error {
		close(blockerStarted)
		<-release
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	select {
	case <-blockerStarted:
	case <-time.After(time.Second):
		t.Fatal("blocker did not start")
	}

	ctx, cancel := context.WithCancel(context.Background())
	var ran atomic.Bool
	errCh := make(chan error, 1)
	go func() {
		errCh <- p.Go(ctx, func(context.Context) error {
			ran.Store(true)
			return nil
		})
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Go did not return after ctx cancel")
	}

	if ran.Load() {
		t.Fatal("expected fn not to run when cancelled while waiting")
	}

	close(release)
}

func TestPool_ctxCancelWhileRunning(t *testing.T) {
	p, err := NewPool(2)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	running := make(chan struct{})
	done := make(chan struct{})

	if err := p.Go(ctx, func(c context.Context) error {
		close(running)
		time.Sleep(100 * time.Millisecond)
		close(done)
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	select {
	case <-running:
	case <-time.After(time.Second):
		t.Fatal("task did not start")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("running task should complete even after ctx cancel")
	}
}

func TestPool_skipsWhenCtxAlreadyCancelled(t *testing.T) {
	p, err := NewPool(1)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var ran atomic.Bool
	err = p.Go(ctx, func(context.Context) error {
		ran.Store(true)
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if ran.Load() {
		t.Fatal("expected fn not to run")
	}
}

func TestPool_multipleTasksComplete(t *testing.T) {
	p, err := NewPool(3)
	if err != nil {
		t.Fatal(err)
	}

	const n = 10
	var wg sync.WaitGroup
	wg.Add(n)
	ctx := context.Background()

	for i := 0; i < n; i++ {
		if err := p.Go(ctx, func(context.Context) error {
			defer wg.Done()
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("not all tasks completed")
	}
}
