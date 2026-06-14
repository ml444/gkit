package cryptox

import (
	"encoding/hex"
	"reflect"
	"testing"
)

func TestAESSIV(t *testing.T) {
	x, err := NewAESSIV(
		[]byte("u8pAdGgDU4Yw59aIFfieNiJNRrmHWYj1"),
	)
	if err != nil {
		panic(err.Error())
	}
	var (
		testString = "test"
		testBytes  = []byte("test")
	)

	s1, err := x.Encrypt(testString)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	plainText, err := x.Decrypt(s1)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	if plainText != testString {
		t.Errorf("plainText: %v, want %v", plainText, testString)
	}
	s2, err := x.Encrypt(testString)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	if s1 != s2 {
		t.Errorf("ciphertext: %v, \n      want %v", s2, s1)
		return
	}

	b1, err := x.Encrypt(testBytes)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	plainText2, err := x.Decrypt(b1)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	if !reflect.DeepEqual(plainText2, testBytes) {
		t.Errorf("plainText: %v, want %v", plainText2, testBytes)
	}
	b2, err := x.Encrypt(testBytes)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	if !reflect.DeepEqual(b1, b2) {
		t.Errorf("ciphertext: %v, \n      want %v", b2, b1)
		return
	}
}

func TestAESSIVRFC5297Vector(t *testing.T) {
	key := mustDecodeHex(t, "fffefdfcfbfaf9f8f7f6f5f4f3f2f1f0f0f1f2f3f4f5f6f7f8f9fafbfcfdfeff")
	ad := mustDecodeHex(t, "101112131415161718191a1b1c1d1e1f2021222324252627")
	plaintext := mustDecodeHex(t, "112233445566778899aabbccddee")
	want := mustDecodeHex(t, "85632d07c6e8f37f950acd320a2ecc9340c02b9690c4dc04daef7f6afe5c")

	x, err := NewAESSIV(key, AESSIVOptWithDataByte(ad))
	if err != nil {
		t.Fatal(err)
	}
	got, err := x.EncryptWithBytes(plaintext)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ciphertext: %x, want %x", got, want)
	}

	decrypted, err := x.DecryptWithBytes(got)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decrypted, plaintext) {
		t.Errorf("plaintext: %x, want %x", decrypted, plaintext)
	}
}

func TestAESSIVRejectsTampering(t *testing.T) {
	x, err := NewAESSIV([]byte("u8pAdGgDU4Yw59aIFfieNiJNRrmHWYj1"))
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := x.EncryptWithBytes([]byte("test"))
	if err != nil {
		t.Fatal(err)
	}
	ciphertext[len(ciphertext)-1] ^= 0x01
	if _, err := x.DecryptWithBytes(ciphertext); err == nil {
		t.Fatal("expected authentication error")
	}
}

func mustDecodeHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func Test_ExampleNewAESSIV(t *testing.T) {
	cryptor, err := NewAESSIV([]byte("12345678901234561234567890123456"))
	if err != nil {
		panic(err)
	}

	ciphertext1, err := cryptor.EncryptWithString("same plaintext")
	if err != nil {
		panic(err)
	}
	ciphertext2, err := cryptor.EncryptWithString("same plaintext")
	if err != nil {
		panic(err)
	}
	plaintext, err := cryptor.DecryptWithString(ciphertext1)
	if err != nil {
		panic(err)
	}
	if ciphertext1 != ciphertext2 {
		t.Errorf("ciphertext1 and ciphertext2 should be the same, got %q and %q", ciphertext1, ciphertext2)
	}
	if plaintext != "same plaintext" {
		t.Errorf("decrypted plaintext should be 'same plaintext', got %q", plaintext)
	}

	// Output:
	// true
	// same plaintext
}
