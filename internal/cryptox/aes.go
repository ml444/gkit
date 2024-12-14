package cryptox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/ml444/gkit/log"
)

type stringer interface {
	EncodeToString([]byte) string
	DecodeString(string) ([]byte, error)
}

type OptFunc func(c *AES)

func OptWithNonceSize(size int) OptFunc {
	return func(c *AES) {
		c.nonceSize = size
	}
}
func OptWithDataByte(data []byte) OptFunc {
	return func(c *AES) {
		c.AdditionalData = data
	}
}

func OptWithStringer(i stringer) OptFunc {
	return func(c *AES) {
		c.stringer = i
	}
}

type AES struct {
	AdditionalData []byte
	key            []byte
	nonceSize      int
	gcm            cipher.AEAD
	stringer       stringer
}

func NewAES(key []byte, opts ...OptFunc) (*AES, error) {
	c := &AES{
		key:       key,
		nonceSize: 12,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.stringer == nil {
		c.stringer = base64.RawStdEncoding
	}
	block, err := aes.NewCipher(c.key)
	if err != nil {
		log.Errorf("NewCipher err: %v\n", err)
		return nil, err
	}
	c.gcm, err = cipher.NewGCMWithNonceSize(block, c.nonceSize)
	if err != nil {
		log.Errorf("NewGCM err: %v\n", err)
		return nil, err
	}
	return c, nil
}

func (x *AES) Encrypt(plaintext any) (any, error) {
	switch v := plaintext.(type) {
	case string:
		return x.EncryptWithString(v)
	case []byte:
		return x.EncryptWithBytes(v)
	default:
		return nil, fmt.Errorf("ciphertext type [%T] not supported", plaintext)
	}
}

func (x *AES) Decrypt(ciphertext any) (any, error) {
	switch v := ciphertext.(type) {
	case string:
		return x.DecryptWithString(v)
	case []byte:
		return x.DecryptWithBytes(v)
	default:
		return nil, fmt.Errorf("ciphertext type [%T] not supported", ciphertext)
	}
}

func (x *AES) EncryptWithString(plaintext string) (string, error) {
	cipherBuf, err := x.EncryptWithBytes([]byte(plaintext))
	if err != nil {
		log.Errorf("Encrypt err: %v", err)
		return "", err
	}
	return x.stringer.EncodeToString(cipherBuf), nil
}

func (x *AES) EncryptWithBytes(plaintext []byte) ([]byte, error) {
	nonce := x.NewNonce()
	ciphertext := x.gcm.Seal(nil, nonce, plaintext, x.AdditionalData)
	return append(nonce, ciphertext...), nil
}

func (x *AES) DecryptWithString(ciphertext string) (string, error) {
	cipherBuf, err := x.stringer.DecodeString(ciphertext)
	if err != nil {
		log.Errorf("DecodeString err: %v\n", err)
		return "", err
	}
	plaintext, err := x.DecryptWithBytes(cipherBuf)
	if err != nil {
		log.Errorf("Decrypt err: %v\n", err)
		return "", err
	}
	return string(plaintext), nil
}

func (x *AES) DecryptWithBytes(cipherBuf []byte) ([]byte, error) {
	plaintext, err := x.gcm.Open(nil, cipherBuf[:x.nonceSize], cipherBuf[x.nonceSize:], x.AdditionalData)
	if err != nil {
		log.Errorf("Decrypt err: %v\n", err)
		return nil, err
	}
	return plaintext, nil
}

func (x *AES) Encode2Str(b []byte) string {
	return x.stringer.EncodeToString(b)
}
func (x *AES) Decode2Byte(s string) ([]byte, error) {
	return x.stringer.DecodeString(s)
}
func (x *AES) NewNonce() []byte {
	return genNonce(x.nonceSize)
}
func (x *AES) NewNonceStr() string {
	return x.stringer.EncodeToString(x.NewNonce())
}
func genNonce(size int) []byte {
	nonce := make([]byte, size)
	_, _ = rand.Read(nonce)
	return nonce
}
