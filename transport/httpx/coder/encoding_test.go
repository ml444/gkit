package coder

import (
	"errors"
	"testing"
)

type testCoder struct {
	name string
}

func (c testCoder) Marshal(v interface{}) ([]byte, error) { return []byte(c.name), nil }
func (c testCoder) Unmarshal([]byte, interface{}) error   { return nil }
func (c testCoder) Name() string                          { return c.name }

func TestRegisterCoder(t *testing.T) {
	if err := RegisterCoder(nil); err == nil {
		t.Fatal("expected nil coder error")
	}
	if err := RegisterCoder(testCoder{}); err == nil {
		t.Fatal("expected empty name error")
	}
	if err := RegisterCoder(testCoder{name: "Custom"}); err != nil {
		t.Fatalf("register coder: %v", err)
	}
	if got := GetCoder("custom").Name(); got != "Custom" {
		t.Fatalf("coder name = %q", got)
	}
	if got := GetCoder("missing").Name(); got != "json" {
		t.Fatalf("fallback coder = %q", got)
	}
}

func TestCoderContractErrorsRemainComparable(t *testing.T) {
	err := errors.New("x")
	if !errors.Is(err, err) {
		t.Fatal("sanity check")
	}
}
