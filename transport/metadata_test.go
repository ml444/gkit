package transport

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

type fakeTransport struct {
	in  MD
	out MD
}

func (fakeTransport) Kind() string     { return "fake" }
func (fakeTransport) Endpoint() string { return "127.0.0.1:0" }
func (fakeTransport) Path() string     { return "/fake.Service/Method" }
func (t fakeTransport) In() MD         { return t.in }
func (t fakeTransport) Out() MD        { return t.out }

func TestContextTransport(t *testing.T) {
	tr := fakeTransport{in: Pairs("trace", "abc"), out: MD{}}
	ctx := ToContext(context.Background(), tr)

	got, ok := FromContext(ctx)
	if !ok {
		t.Fatal("expected transport in context")
	}
	if got.Kind() != "fake" || got.Endpoint() == "" || got.Path() == "" {
		t.Fatalf("unexpected transport: %#v", got)
	}

	if _, ok := FromContext(context.Background()); ok {
		t.Fatal("expected no transport in empty context")
	}
}

func TestMetadataOperations(t *testing.T) {
	md := New(map[string][]string{"a": {"1", "2"}}, map[string][]string{"b": {"3"}})
	md.Append("a", "4")
	md.Append("empty")
	md.Set("c", "5", "6")
	md.Set("empty-set")

	if got := md.GetFirst("a"); got != "1" {
		t.Fatalf("first a = %q", got)
	}
	if got := md.Get("a"); !reflect.DeepEqual(got, []string{"1", "2", "4"}) {
		t.Fatalf("a = %#v", got)
	}
	if got := md.Get("c"); !reflect.DeepEqual(got, []string{"5", "6"}) {
		t.Fatalf("c = %#v", got)
	}
	if md.Len() != 3 {
		t.Fatalf("len = %d", md.Len())
	}

	seen := map[string]bool{}
	md.Range(func(k string, _ []string) bool {
		seen[k] = true
		return k != "b"
	})
	if len(md.Keys()) != 3 {
		t.Fatalf("keys = %#v", md.Keys())
	}

	cp := md.Copy()
	cp.Append("a", "copy")
	if reflect.DeepEqual(cp.Get("a"), md.Get("a")) {
		t.Fatal("copy should not share value slices")
	}

	merged := Merge(Pairs("x", "1"), Pairs("x", "2", "y", "3"))
	if got := merged.Get("x"); !reflect.DeepEqual(got, []string{"1", "2"}) {
		t.Fatalf("merged x = %#v", got)
	}
	md.Delete("b")
	if md.GetFirst("b") != "" {
		t.Fatal("expected b to be deleted")
	}
}

func TestMetadataKeysAreLowercase(t *testing.T) {
	md := New(map[string][]string{
		"Content-Type": {"application/json"},
		"X-Trace-ID":   {"trace-1"},
	})
	md.Append("X-Trace-ID", "trace-2")
	md.Set("X-Request-ID", "req-1")

	if got := md.GetFirst("content-type"); got != "application/json" {
		t.Fatalf("content-type = %q", got)
	}
	if got := md.Get("x-trace-id"); !reflect.DeepEqual(got, []string{"trace-1", "trace-2"}) {
		t.Fatalf("x-trace-id = %#v", got)
	}
	if got := md.GetFirst("X-REQUEST-ID"); got != "req-1" {
		t.Fatalf("x-request-id = %q", got)
	}
	for _, key := range md.Keys() {
		if key != strings.ToLower(key) {
			t.Fatalf("key %q is not lowercase", key)
		}
	}

	raw := MD{"Mixed-Key": {"1"}, "mixed-key": {"2"}}
	if raw.Len() != 1 {
		t.Fatalf("normalized len = %d", raw.Len())
	}
	if got := raw.Get("MIXED-KEY"); !hasValues(got, "2") {
		t.Fatalf("mixed get = %#v", got)
	}
	raw.Delete("MIXED-KEY")
	if raw.Len() != 0 {
		t.Fatalf("delete should remove all case variants: %#v", raw)
	}

	merged := Merge(Pairs("X-A", "1"), MD{"x-a": {"2"}, "X-B": {"3"}})
	if got := merged.Get("x-a"); !reflect.DeepEqual(got, []string{"1", "2"}) {
		t.Fatalf("merged x-a = %#v", got)
	}
	if got := merged.Copy().GetFirst("x-b"); got != "3" {
		t.Fatalf("copy x-b = %q", got)
	}
}

func hasValues(got []string, want ...string) bool {
	if len(got) != len(want) {
		return false
	}
	seen := make(map[string]int, len(got))
	for _, v := range got {
		seen[v]++
	}
	for _, v := range want {
		if seen[v] == 0 {
			return false
		}
		seen[v]--
	}
	return true
}

func TestPairsPanicsOnOddInput(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	_ = Pairs("odd")
}
