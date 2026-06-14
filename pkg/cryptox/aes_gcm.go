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
	case []string:
		result := []string{}
		for _, s := range v {
			vv, err := x.EncryptWithString(s)
			if err != nil {
				return result, err
			}
			result = append(result, string(vv))
		}
		return result, nil
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
	case []string:
		var result []string
		for _, s := range v {
			vv, err := x.DecryptWithString(s)
			if err != nil {
				return result, err
			}
			result = append(result, string(vv))
		}
		return result, nil
	default:
		return nil, fmt.Errorf("ciphertext type [%T] not supported", ciphertext)
	}
}

func (x *AES) EncryptWithString(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	cipherBuf, err := x.EncryptWithBytes([]byte(plaintext))
	if err != nil {
		log.Errorf("Encrypt err: %v", err)
		return "", err
	}
	return x.encoder.EncodeToString(cipherBuf), nil
}

func (x *AES) EncryptWithBytes(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return []byte{}, nil
	}
	nonce, err := genNonce(x.nonceSize)
	if err != nil {
		log.Errorf("generate nonce err: %v", err)
		return nil, err
	}
	ciphertext := x.gcm.Seal(nil, nonce, plaintext, x.AdditionalData)
	return append(nonce, ciphertext...), nil
}

func (x *AES) DecryptWithString(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
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
	if len(cipherBuf) == 0 {
		return []byte{}, nil
	}
	if len(cipherBuf) < x.nonceSize {
		log.Errorf("cipherBuf len %d less than nonce size %d\n", len(cipherBuf), x.nonceSize)
		return nil, fmt.Errorf("cipherBuf len %d less than nonce size %d", len(cipherBuf), x.nonceSize)
	}
	plaintext, err := x.gcm.Open(nil, cipherBuf[:x.nonceSize], cipherBuf[x.nonceSize:], x.AdditionalData)
	if err != nil {
		log.Errorf("Decrypt err: %v\n", err)
		return nil, err
	}
	return plaintext, nil
}

func (x *AES) NewNonce() []byte {
	nonce, err := genNonce(x.nonceSize)
	if err != nil {
		log.Errorf("generate nonce err: %v", err)
		return nil
	}
	return nonce
}

func (x *AES) NewNonceStr() string {
	return x.encoder.EncodeToString(x.NewNonce())
}

// genNonce returns a cryptographically secure random nonce.
// It returns an error if the system CSPRNG fails, so callers never
// accidentally encrypt with a predictable/zero nonce.
func genNonce(size int) ([]byte, error) {
	nonce := make([]byte, size)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}

type OptFunc func(c *AES)

// AESOptWithNonceSize sets the nonce size for the AES encryption.
// The default nonce size is 12 bytes, which is the recommended size.
// If you need to use a different nonce size, you can set it using this option.
func AESOptWithNonceSize(size int) OptFunc {
	return func(c *AES) {
		c.nonceSize = size
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
