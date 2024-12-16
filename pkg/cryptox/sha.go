package cryptox

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
)

func Sha1(dst string) string {
	h := sha1.New()
	_, _ = h.Write([]byte(dst))
	return hex.EncodeToString(h.Sum(nil))
}

func Sha1WithSalt(dst string, salt string) string {
	h := sha1.New()
	_, _ = h.Write([]byte(dst + salt))
	return hex.EncodeToString(h.Sum(nil))
}

func Sha256(dst string) string {
	h := sha256.New()
	_, _ = h.Write([]byte(dst))
	return hex.EncodeToString(h.Sum(nil))
}

func Sha256WithSalt(dst string, salt string) string {
	h := sha256.New()
	_, _ = h.Write([]byte(dst + salt))
	return hex.EncodeToString(h.Sum(nil))
}

func Sha512(dst string) string {
	h := sha512.New()
	_, _ = h.Write([]byte(dst))
	return hex.EncodeToString(h.Sum(nil))
}

func Sha512WithSalt(dst string, salt string) string {
	h := sha512.New()
	_, _ = h.Write([]byte(dst + salt))
	return hex.EncodeToString(h.Sum(nil))
}
