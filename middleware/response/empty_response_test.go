package response

import (
	"context"
	"testing"
)

type emptyMsg struct{}

func TestReplaceEmptyResponse_NilPointer(t *testing.T) {
	mw := ReplaceEmptyResponse("ok")
	var rsp *emptyMsg
	out, err := mw(func(ctx context.Context, req interface{}) (interface{}, error) {
		return rsp, nil
	})(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "ok" {
		t.Fatalf("out = %v", out)
	}
}

func TestReplaceEmptyResponse_NonPointerNoPanic(t *testing.T) {
	mw := ReplaceEmptyResponse("ok")
	out, err := mw(func(ctx context.Context, req interface{}) (interface{}, error) {
		return 42, nil
	})(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != 42 {
		t.Fatalf("out = %v", out)
	}
}
