package main

import (
	"testing"
)

func TestSortedCommons(t *testing.T) {
	got := sortedCommons(map[string]string{
		"jsonMarshal": "json-func",
		"bytesMarshal": "bytes-func",
		"datetime": "date-func",
	})
	want := []string{"bytes-func", "date-func", "json-func"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sortedCommons[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestAppendImportsDedupes(t *testing.T) {
	var imports []string
	appendImports(&imports, "fmt", "fmt", "encoding/json")
	appendImports(&imports, "encoding/json", "time")
	if len(imports) != 3 {
		t.Fatalf("imports = %v, want 3 unique entries", imports)
	}
}
