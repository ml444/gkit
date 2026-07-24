package crypto

import (
	"strings"
	"testing"
)

func testRing(t *testing.T) KeyRing {
	t.Helper()
	key := []byte("0123456789abcdef0123456789abcdef") // 32 bytes
	return KeyRing{ActiveID: "1", Keys: map[string][]byte{
		"1": key,
		"0": []byte("abcdef0123456789abcdef0123456789"),
	}}
}

func TestPrimitiveStorageRoundTrip(t *testing.T) {
	p := NewPrimitive()
	r := testRing(t)
	active, _ := r.Active()
	enc, err := p.Encrypt(ModeStorage, active, "secret")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(enc, "v1:k1:") {
		t.Fatalf("envelope = %q", enc)
	}
	enc2, _ := p.Encrypt(ModeStorage, active, "secret")
	if enc == enc2 {
		t.Fatal("storage mode must be non-deterministic")
	}
	plain, err := p.Decrypt(ModeStorage, r, enc)
	if err != nil || plain != "secret" {
		t.Fatalf("decrypt = %q, %v", plain, err)
	}
}

func TestPrimitiveSearchableDeterministicAndOldKey(t *testing.T) {
	p := NewPrimitive()
	r := testRing(t)
	k0, _ := r.Get("0")
	enc, err := p.Encrypt(ModeSearchable, k0, "phone")
	if err != nil {
		t.Fatal(err)
	}
	enc2, _ := p.Encrypt(ModeSearchable, k0, "phone")
	if enc != enc2 {
		t.Fatal("searchable must be deterministic")
	}
	plain, err := p.Decrypt(ModeSearchable, r, enc)
	if err != nil || plain != "phone" {
		t.Fatalf("decrypt old key = %q, %v", plain, err)
	}
}
