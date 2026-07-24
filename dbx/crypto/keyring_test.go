package crypto

import (
	"errors"
	"testing"
)

func TestKeyRingActiveAndGet(t *testing.T) {
	r := KeyRing{
		ActiveID: "1",
		Keys: map[string][]byte{
			"1": []byte("0123456789abcdef0123456789abcdef"),
			"0": []byte("abcdef0123456789abcdef0123456789"),
		},
	}
	if err := r.Validate(); err != nil {
		t.Fatal(err)
	}
	active, err := r.Active()
	if err != nil || active.ID != "1" {
		t.Fatalf("active = %#v, %v", active, err)
	}
	old, err := r.Get("0")
	if err != nil || len(old.Material) != 32 {
		t.Fatalf("old = %#v, %v", old, err)
	}
	if _, err := r.Get("9"); !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("want ErrUnknownKey, got %v", err)
	}
}

func TestKeyRingValidate(t *testing.T) {
	if err := (KeyRing{}).Validate(); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("empty: %v", err)
	}
}
