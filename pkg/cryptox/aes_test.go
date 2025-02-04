package cryptox

import (
	"reflect"
	"testing"
)

func TestAES(t *testing.T) {
	x, err := NewAES(
		[]byte("u8pAdGgDU4Yw59aIFfieNiJNRrmHWYj1"),
		// AESOptWithDataByte([]byte("test")),
	)
	if err != nil {
		panic(err.Error())
	}
	var (
		testString = "test"
		testBytes  = []byte("test")
	)
	// test string
	s, err := x.Encrypt(testString)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	t.Logf("ciphertext: %s\n", s.(string))
	plainText, err := x.Decrypt(s)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	if plainText != testString {
		t.Errorf("plainText: %v, want %v", plainText, testString)
	}

	// test bytes
	b, err := x.Encrypt(testBytes)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	plainText2, err := x.Decrypt(b)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	if !reflect.DeepEqual(plainText2, testBytes) {
		t.Errorf("plainText: %v, want %v", plainText2, testBytes)
	}
}

func TestAESWithFixedNonce(t *testing.T) {
	t.Log(len([]byte("fixedNonce12")))
	x, err := NewAES(
		[]byte("u8pAdGgDU4Yw59aIFfieNiJNRrmHWYj1"),
		// AESOptWithDataByte([]byte("test")),
		AESOptWithFixedNonce([]byte("fixedNonce12")),
	)
	if err != nil {
		panic(err.Error())
	}
	var (
		testString = "test"
		testBytes  = []byte("test")
	)
	// test string
	s1, err := x.Encrypt(testString)
	if err != nil {
		t.Errorf("err: %v\n", err)
		return
	}
	t.Logf("ciphertext: %s\n", s1.(string))
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

	// test bytes
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
