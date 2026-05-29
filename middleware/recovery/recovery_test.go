package recovery

import (
	"context"
	"errors"
	"testing"

	"github.com/ml444/gkit/errorx"
)

func TestRecovery_CatchesPanic(t *testing.T) {
	mw := Recovery()
	h := mw(func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("boom")
	})
	_, err := h(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var ex *errorx.Error
	if !errors.As(err, &ex) {
		t.Fatalf("expected errorx.Error, got %T", err)
	}
}
