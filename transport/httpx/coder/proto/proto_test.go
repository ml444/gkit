package proto

import (
	"testing"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestCoderProto(t *testing.T) {
	c := GetCoder()
	if c.Name() != Name {
		t.Fatalf("name = %q", c.Name())
	}
	data, err := c.Marshal(wrapperspb.String("hello"))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out wrapperspb.StringValue
	if err := c.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Value != "hello" {
		t.Fatalf("value = %q", out.Value)
	}
	ptr := wrapperspb.String("")
	ptrPtr := &ptr
	if err := c.Unmarshal(data, &ptrPtr); err != nil || ptrPtr == nil || (*ptrPtr).Value != "hello" {
		t.Fatalf("nested pointer unmarshal = %#v, %v", ptrPtr, err)
	}
	if err := c.Unmarshal(data, struct{}{}); err == nil {
		t.Fatal("expected non-proto error")
	}
}
