package cryptox

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"github.com/ml444/gkit/log"
)

const aesSIVTagSize = aes.BlockSize

// AESSIV implements AES-SIV as a deterministic AEAD.
//
// Use this when the same plaintext must produce the same ciphertext, such as
// exact-match lookup fields. Unlike AES-GCM with a fixed nonce, AES-SIV is
// designed for deterministic encryption.
//
// Keys must be double-length AES keys:
//   - 32 bytes for AES-128-SIV
//   - 48 bytes for AES-192-SIV
//   - 64 bytes for AES-256-SIV
//
// Go's standard library does not provide AES-GCM-SIV. If AES-GCM-SIV is
// required, use a vetted implementation of that mode instead of reusing a GCM
// nonce. For deterministic encryption in this package, prefer NewAESSIV.
type AESSIV struct {
	AdditionalData []byte
	key            []byte
	cmacBlock      cipher.Block
	ctrBlock       cipher.Block
	encoder        encoder
}

func NewAESSIV(key []byte, opts ...SIVOptFunc) (*AESSIV, error) {
	if len(key) != 32 && len(key) != 48 && len(key) != 64 {
		return nil, fmt.Errorf("cryptox: AES-SIV key size must be 32, 48, or 64 bytes, got %d", len(key))
	}

	x := &AESSIV{
		key:     key,
		encoder: base64.RawStdEncoding,
	}
	for _, opt := range opts {
		opt(x)
	}
	if x.encoder == nil {
		x.encoder = base64.RawStdEncoding
	}

	half := len(key) / 2
	var err error
	x.cmacBlock, err = aes.NewCipher(key[:half])
	if err != nil {
		log.Errorf("NewCipher cmac err: %v\n", err)
		return nil, err
	}
	x.ctrBlock, err = aes.NewCipher(key[half:])
	if err != nil {
		log.Errorf("NewCipher ctr err: %v\n", err)
		return nil, err
	}
	return x, nil
}

func (x *AESSIV) Encrypt(plaintext any) (any, error) {
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
			result = append(result, vv)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("plaintext type [%T] not supported", plaintext)
	}
}

func (x *AESSIV) Decrypt(ciphertext any) (any, error) {
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
			result = append(result, vv)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("ciphertext type [%T] not supported", ciphertext)
	}
}

func (x *AESSIV) EncryptWithString(plaintext string) (string, error) {
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

func (x *AESSIV) EncryptWithBytes(plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return []byte{}, nil
	}

	tag := x.s2v(plaintext)
	iv := aesSIVCTRIV(tag)
	ciphertext := make([]byte, len(plaintext))
	cipher.NewCTR(x.ctrBlock, iv).XORKeyStream(ciphertext, plaintext)

	out := make([]byte, 0, aesSIVTagSize+len(ciphertext))
	out = append(out, tag...)
	out = append(out, ciphertext...)
	return out, nil
}

func (x *AESSIV) DecryptWithString(ciphertext string) (string, error) {
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

func (x *AESSIV) DecryptWithBytes(cipherBuf []byte) ([]byte, error) {
	if len(cipherBuf) == 0 {
		return []byte{}, nil
	}
	if len(cipherBuf) < aesSIVTagSize {
		return nil, fmt.Errorf("cipherBuf len %d less than AES-SIV tag size %d", len(cipherBuf), aesSIVTagSize)
	}

	tag := cipherBuf[:aesSIVTagSize]
	ciphertext := cipherBuf[aesSIVTagSize:]
	iv := aesSIVCTRIV(tag)
	plaintext := make([]byte, len(ciphertext))
	cipher.NewCTR(x.ctrBlock, iv).XORKeyStream(plaintext, ciphertext)

	wantTag := x.s2v(plaintext)
	if subtle.ConstantTimeCompare(tag, wantTag) != 1 {
		return nil, fmt.Errorf("cryptox: AES-SIV authentication failed")
	}
	return plaintext, nil
}

func (x *AESSIV) s2v(plaintext []byte) []byte {
	var zero [aes.BlockSize]byte
	d := aesCMAC(x.cmacBlock, zero[:])

	if len(x.AdditionalData) > 0 {
		d = aesSIVDbl(d)
		adMac := aesCMAC(x.cmacBlock, x.AdditionalData)
		aesSIVXOR(d, adMac)
	}

	var t []byte
	if len(plaintext) >= aes.BlockSize {
		t = append([]byte(nil), plaintext...)
		aesSIVXOREnd(t, d)
	} else {
		t = aesSIVDbl(d)
		padded := aesSIVPad(plaintext)
		aesSIVXOR(t, padded)
	}
	return aesCMAC(x.cmacBlock, t)
}

func aesCMAC(block cipher.Block, message []byte) []byte {
	k1, k2 := aesCMACSubkeys(block)
	n := (len(message) + aes.BlockSize - 1) / aes.BlockSize
	if n == 0 {
		n = 1
	}

	last := make([]byte, aes.BlockSize)
	complete := len(message) > 0 && len(message)%aes.BlockSize == 0
	if complete {
		copy(last, message[(n-1)*aes.BlockSize:n*aes.BlockSize])
		aesSIVXOR(last, k1)
	} else {
		start := (n - 1) * aes.BlockSize
		copy(last, aesSIVPad(message[start:]))
		aesSIVXOR(last, k2)
	}

	x := make([]byte, aes.BlockSize)
	y := make([]byte, aes.BlockSize)
	for i := 0; i < n-1; i++ {
		copy(y, message[i*aes.BlockSize:(i+1)*aes.BlockSize])
		aesSIVXOR(y, x)
		block.Encrypt(x, y)
	}
	aesSIVXOR(last, x)
	block.Encrypt(x, last)
	return x
}

func aesCMACSubkeys(block cipher.Block) ([]byte, []byte) {
	var zero [aes.BlockSize]byte
	l := make([]byte, aes.BlockSize)
	block.Encrypt(l, zero[:])
	k1 := aesSIVDbl(l)
	k2 := aesSIVDbl(k1)
	return k1, k2
}

func aesSIVDbl(in []byte) []byte {
	out := make([]byte, aes.BlockSize)
	var carry byte
	for i := aes.BlockSize - 1; i >= 0; i-- {
		nextCarry := in[i] >> 7
		out[i] = (in[i] << 1) | carry
		carry = nextCarry
	}
	if carry != 0 {
		out[aes.BlockSize-1] ^= 0x87
	}
	return out
}

func aesSIVPad(in []byte) []byte {
	out := make([]byte, aes.BlockSize)
	copy(out, in)
	out[len(in)] = 0x80
	return out
}

func aesSIVXOR(dst, src []byte) {
	for i := range dst {
		dst[i] ^= src[i]
	}
}

func aesSIVXOREnd(dst, src []byte) {
	offset := len(dst) - len(src)
	for i := range src {
		dst[offset+i] ^= src[i]
	}
}

func aesSIVCTRIV(tag []byte) []byte {
	iv := bytes.Clone(tag)
	iv[8] &= 0x7f
	iv[12] &= 0x7f
	return iv
}

type SIVOptFunc func(c *AESSIV)

func AESSIVOptWithDataByte(data []byte) SIVOptFunc {
	return func(c *AESSIV) {
		c.AdditionalData = data
	}
}

func AESSIVOptWithEncoder(i encoder) SIVOptFunc {
	return func(c *AESSIV) {
		c.encoder = i
	}
}
