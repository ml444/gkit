package crypto

import (
	"errors"
	"strings"
	"testing"

	"github.com/ml444/gkit/pkg/cryptox"
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

func TestPrimitiveErrorBranches(t *testing.T) {
	p := NewPrimitive()
	r := testRing(t)
	active, _ := r.Active()
	if enc, err := p.Encrypt(ModeStorage, active, ""); err != nil || enc != "" {
		t.Fatalf("empty encrypt = %q, %v", enc, err)
	}
	if _, err := p.Encrypt(Mode(99), active, "x"); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("invalid encrypt mode: %v", err)
	}
	badKey := Key{ID: "bad", Material: []byte("short")}
	if _, err := p.Encrypt(ModeStorage, badKey, "x"); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("bad aes key: %v", err)
	}
	if _, err := p.Encrypt(ModeSearchable, badKey, "x"); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("bad siv key: %v", err)
	}
	injected := NewPrimitive()
	wantEncryptErr := errors.New("encrypt failed")
	injected.encryptAES = func(*cryptox.AES, []byte) ([]byte, error) {
		return nil, wantEncryptErr
	}
	if _, err := injected.Encrypt(ModeStorage, active, "x"); !errors.Is(err, wantEncryptErr) {
		t.Fatalf("storage encrypt error = %v", err)
	}
	injected = NewPrimitive()
	injected.encryptSIV = func(*cryptox.AESSIV, []byte) ([]byte, error) {
		return nil, wantEncryptErr
	}
	if _, err := injected.Encrypt(ModeSearchable, active, "x"); !errors.Is(err, wantEncryptErr) {
		t.Fatalf("searchable encrypt error = %v", err)
	}
	if plain, err := p.Decrypt(ModeStorage, r, ""); err != nil || plain != "" {
		t.Fatalf("empty decrypt = %q, %v", plain, err)
	}
	if _, err := p.Decrypt(ModeStorage, r, "not-envelope"); !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("bad envelope: %v", err)
	}
	if _, err := p.Decrypt(ModeStorage, r, "v1:k9:abc"); !errors.Is(err, ErrUnknownKey) {
		t.Fatalf("unknown key: %v", err)
	}
	if _, err := p.Decrypt(ModeStorage, r, "v1:k1:%%"); !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("bad payload: %v", err)
	}
	searchable, err := p.Encrypt(ModeSearchable, active, "x")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Decrypt(ModeStorage, r, searchable); !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("wrong decrypt mode: %v", err)
	}
	storage, err := p.Encrypt(ModeStorage, active, "x")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := p.Decrypt(ModeSearchable, r, storage); !errors.Is(err, ErrDecryptFailed) {
		t.Fatalf("wrong searchable mode: %v", err)
	}
	if _, err := p.Decrypt(Mode(99), r, storage); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("invalid decrypt mode: %v", err)
	}
	badRing := KeyRing{ActiveID: "bad", Keys: map[string][]byte{"bad": []byte("short")}}
	badCiphertext := FormatEnvelope(Envelope{Version: EnvelopeVersion, KeyID: "bad", Payload: "abc"})
	if _, err := NewPrimitive().Decrypt(ModeStorage, badRing, badCiphertext); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("bad aes decrypt key: %v", err)
	}
	if _, err := NewPrimitive().Decrypt(ModeSearchable, badRing, badCiphertext); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("bad siv decrypt key: %v", err)
	}
}
