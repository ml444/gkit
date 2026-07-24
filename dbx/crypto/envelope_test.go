package crypto

import "testing"

func TestEnvelopeRoundTrip(t *testing.T) {
	in := Envelope{Version: 1, KeyID: "2", Payload: "abc+/=_"}
	s := FormatEnvelope(in)
	if s != "v1:k2:abc+/=_" {
		t.Fatalf("format = %q", s)
	}
	out, err := ParseEnvelope(s)
	if err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Fatalf("parse = %#v", out)
	}
}

func TestParseEnvelopeInvalid(t *testing.T) {
	for _, s := range []string{"", "nope", "v1:k", "vX:k1:p", "v1:x1:p"} {
		if _, err := ParseEnvelope(s); err == nil {
			t.Fatalf("expected error for %q", s)
		}
	}
}
