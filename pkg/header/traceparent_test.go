package header

import "testing"

func TestParseTraceparentValid(t *testing.T) {
	tp := "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
	ti, ok := ParseTraceparent(tp)
	if !ok {
		t.Fatal("expected valid traceparent")
	}
	if ti.TraceID != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("trace_id = %q", ti.TraceID)
	}
	if ti.SpanID != "00f067aa0ba902b7" {
		t.Fatalf("span_id = %q", ti.SpanID)
	}
}

func TestParseTraceparentInvalid(t *testing.T) {
	cases := []string{
		"",
		"00-abc",
		"01-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
		"00-00000000000000000000000000000000-00f067aa0ba902b7-01",
		"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-zz",
	}
	for _, c := range cases {
		if _, ok := ParseTraceparent(c); ok {
			t.Fatalf("expected invalid traceparent %q", c)
		}
	}
}
