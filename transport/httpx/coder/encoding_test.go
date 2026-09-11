package coder

import (
	"errors"
	"fmt"
	"sync"
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

func TestLookupAndSnapshot(t *testing.T) {
	if err := RegisterCoder((*testCoder)(nil)); err == nil {
		t.Fatal("accepted typed nil codec")
	}
	if c, ok := LookupCoder("not-registered"); ok || c != nil {
		t.Fatalf("strict lookup unexpectedly fell back: %v, %v", c, ok)
	}
	if c, ok := LookupCoder("JSON"); !ok || c.Name() != "json" {
		t.Fatalf("case insensitive lookup: %v, %v", c, ok)
	}
	snapshot := Snapshot()
	delete(snapshot, "json")
	snapshot["snapshot-only"] = testCoder{name: "snapshot-only"}
	if _, ok := LookupCoder("json"); !ok {
		t.Fatal("snapshot mutation changed registry")
	}
	if _, ok := LookupCoder("snapshot-only"); ok {
		t.Fatal("snapshot insertion changed registry")
	}
	if err := RegisterCoder(testCoder{name: "registered-after-snapshot"}); err != nil {
		t.Fatal(err)
	}
	if _, ok := snapshot["registered-after-snapshot"]; ok {
		t.Fatal("registry mutation changed snapshot")
	}
}

func TestConcurrentRegistryAccess(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				if err := RegisterCoder(testCoder{name: "concurrent"}); err != nil {
					t.Error(err)
				}
				if c, ok := LookupCoder("concurrent"); !ok || c.Name() != "concurrent" {
					t.Errorf("lookup: %v, %v", c, ok)
				}
				if c := GetCoder("unknown-concurrent"); c.Name() != "json" {
					t.Error("lost JSON fallback")
				}
				_ = Snapshot()
			}
		}()
	}
	wg.Wait()
}

func TestRegisterDoesNotMutatePublishedVersion(t *testing.T) {
	const name = "immutable-version"
	first, second := &testCoder{name: name}, &testCoder{name: name}
	if err := RegisterCoder(first); err != nil {
		t.Fatal(err)
	}
	previous := registeredCoders.Load()
	if err := RegisterCoder(second); err != nil {
		t.Fatal(err)
	}
	if previous.codecs[name] != first {
		t.Fatal("registration mutated a previously published version")
	}
	if GetCoder(name) != second {
		t.Fatal("readers did not observe the replacement")
	}
}

func TestConcurrentRegistrationsDoNotLoseUpdates(t *testing.T) {
	const writers = 32
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			for j := 0; j < 4; j++ {
				name := fmt.Sprintf("distinct-writer-%d-%d", i, j)
				if err := RegisterCoder(testCoder{name: name}); err != nil {
					t.Error(err)
				}
			}
		}(i)
	}
	close(start)
	wg.Wait()
	snapshot := Snapshot()
	for i := 0; i < writers; i++ {
		for j := 0; j < 4; j++ {
			name := fmt.Sprintf("distinct-writer-%d-%d", i, j)
			if c, ok := snapshot[name]; !ok || c.Name() != name {
				t.Errorf("lost registration %q", name)
			}
		}
	}
}
