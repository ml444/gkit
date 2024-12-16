package cryptox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/ml444/gkit/log"
)

type AES struct {
	AdditionalData []byte
	key            []byte
	FixedNonce     []byte
	nonceSize      int
	gcm            cipher.AEAD
	encoder        encoder
}

func NewAES(key []byte, opts ...OptFunc) (*AES, error) {
	c := &AES{
		key:       key,
		nonceSize: 12,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.encoder == nil {
		c.encoder = base64.RawStdEncoding
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
	return x.encoder.EncodeToString(cipherBuf), nil
}

func (x *AES) EncryptWithBytes(plaintext []byte) ([]byte, error) {
	nonce := x.FixedNonce
	if nonce == nil {
		nonce = x.NewNonce()
	}
	ciphertext := x.gcm.Seal(nil, nonce, plaintext, x.AdditionalData)
	return append(nonce, ciphertext...), nil
}

func (x *AES) DecryptWithString(ciphertext string) (string, error) {
	cipherBuf, err := x.encoder.DecodeString(ciphertext)
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

func (x *AES) NewNonce() []byte {
	return genNonce(x.nonceSize)
}
func (x *AES) NewNonceStr() string {
	return x.encoder.EncodeToString(x.NewNonce())
}
func genNonce(size int) []byte {
	nonce := make([]byte, size)
	_, _ = rand.Read(nonce)
	return nonce
}

type OptFunc func(c *AES)

// AESOptWithNonceSize sets the nonce size for the AES encryption.
// The default nonce size is 12 bytes, which is the recommended size.
// If you need to use a different nonce size, you can set it using this option.
// NOTE: If you set a fixed nonce, this setting will be invalid.
func AESOptWithNonceSize(size int) OptFunc {
	return func(c *AES) {
		if c.FixedNonce != nil {
			return
		}
		c.nonceSize = size
	}
}

// AESOptWithFixedNonce sets the fixed nonce for the AES encryption.
// This setting enables the same string to be encrypted with the same result.
// This is needed in some business scenarios. However, it should be used with
// caution as it can lead to security vulnerabilities.
// If not sets this option will generate a new nonce for each encryption.
func AESOptWithFixedNonce(nonce []byte) OptFunc {
	return func(c *AES) {
		c.FixedNonce = nonce
		c.nonceSize = len(nonce)
	}
}

func AESOptWithDataByte(data []byte) OptFunc {
	return func(c *AES) {
		c.AdditionalData = data
	}
}

func AESOptWithEncoder(i encoder) OptFunc {
	return func(c *AES) {
		c.encoder = i
	}
}
