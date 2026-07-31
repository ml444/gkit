package cryptox

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/ml444/gkit/log"
)

const aesSIVTagSize = aes.BlockSize

var blockPool = sync.Pool{
	New: func() any {
		return new([aesSIVTagSize]byte)
	},
}

func getBlock() *[aesSIVTagSize]byte {
	return blockPool.Get().(*[aesSIVTagSize]byte)
}

func putBlock(b *[aesSIVTagSize]byte) {
	// Go 1.21+ 可以直接使用 clear(b[:])
	for i := range b {
		b[i] = 0
	}
	blockPool.Put(b)
}

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

	// 预计算的 CMAC 子密钥
	k1 [aesSIVTagSize]byte
	k2 [aesSIVTagSize]byte
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
	// 预计算 k1 和 k2 缓存起来
	var zero [aesSIVTagSize]byte
	lPtr := getBlock()
	defer putBlock(lPtr)
	l := lPtr[:]
	
	x.cmacBlock.Encrypt(l, zero[:])
	aesSIVDbl(x.k1[:], l)
	aesSIVDbl(x.k2[:], x.k1[:])
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
	var zero [aesSIVTagSize]byte

	dPtr := getBlock()
	defer putBlock(dPtr)
	d := dPtr[:]

	aesCMAC(d, x.cmacBlock, zero[:], x.k1[:], x.k2[:], nil)

	if len(x.AdditionalData) > 0 {
		aesSIVDbl(d, d) // 原位安全

		adMacPtr := getBlock()
		defer putBlock(adMacPtr)
		adMac := adMacPtr[:]

		aesCMAC(adMac, x.cmacBlock, x.AdditionalData, x.k1[:], x.k2[:], nil)
		aesSIVXOR(d, adMac)
	}

	// out 是本次 s2v 唯一的堆分配，作为 tag 返回外部
	out := make([]byte, aesSIVTagSize) 

	if len(plaintext) >= aesSIVTagSize {
		// 终极性能：直接将 plaintext 丢进去，靠底层流式处理 d，消灭大对象拷贝
		aesCMAC(out, x.cmacBlock, plaintext, x.k1[:], x.k2[:], d)
	} else {
		tPtr := getBlock()
		defer putBlock(tPtr)
		t := tPtr[:]
		aesSIVDbl(t, d)

		paddedPtr := getBlock()
		defer putBlock(paddedPtr)
		padded := paddedPtr[:]
		aesSIVPad(padded, plaintext)

		aesSIVXOR(t, padded)
		aesCMAC(out, x.cmacBlock, t, x.k1[:], x.k2[:], nil)
	}

	return out
}
func aesCMAC(dst []byte, block cipher.Block, msg []byte, k1, k2 []byte, xorEnd []byte) {
	n := (len(msg) + aesSIVTagSize - 1) / aesSIVTagSize
	if n == 0 {
		n = 1
	}

	for i := range dst {
		dst[i] = 0 // 初始化结果缓冲 x
	}

	yPtr := getBlock()
	defer putBlock(yPtr)
	y := yPtr[:]

	offsetInMsg := len(msg) - len(xorEnd)

	// 1. 处理前 n-1 块
	for i := 0; i < n-1; i++ {
		start := i * aesSIVTagSize
		end := start + aesSIVTagSize
		copy(y, msg[start:end])

		// 流式应用 xorEnd (无分配处理跨 Block 边界)
		if len(xorEnd) > 0 && end > offsetInMsg {
			overlapStart := start
			if offsetInMsg > overlapStart {
				overlapStart = offsetInMsg
			}
			for j := overlapStart; j < end; j++ {
				y[j-start] ^= xorEnd[j-offsetInMsg]
			}
		}

		aesSIVXOR(dst, y) // dst 当作原代码里的 x
		block.Encrypt(dst, dst)
	}

	// 2. 处理最后一块
	lastPtr := getBlock()
	defer putBlock(lastPtr)
	last := lastPtr[:]

	complete := len(msg) > 0 && len(msg)%aesSIVTagSize == 0
	start := (n - 1) * aesSIVTagSize
	if complete {
		copy(last, msg[start:])
	} else {
		aesSIVPad(last, msg[start:])
	}

	// 对最后一块应用 xorEnd
	if len(xorEnd) > 0 {
		overlapStart := start
		if offsetInMsg > overlapStart {
			overlapStart = offsetInMsg
		}
		for j := overlapStart; j < len(msg); j++ {
			last[j-start] ^= xorEnd[j-offsetInMsg]
		}
	}

	// 组合 k1 或 k2
	if complete {
		aesSIVXOR(last, k1)
	} else {
		aesSIVXOR(last, k2)
	}

	aesSIVXOR(dst, last)
	block.Encrypt(dst, dst)
}

func aesSIVDbl(out, in []byte) {
	// out := make([]byte, aesSIVTagSize)
	var carry byte
	for i := aesSIVTagSize - 1; i >= 0; i-- {
		nextCarry := in[i] >> 7
		out[i] = (in[i] << 1) | carry
		carry = nextCarry
	}
	if carry != 0 {
		out[aesSIVTagSize-1] ^= 0x87
	}
}

func aesSIVPad(out, in []byte) {
	// out := make([]byte, aesSIVTagSize)
	copy(out, in)
	out[len(in)] = 0x80
	for i := len(in) + 1; i < aesSIVTagSize; i++ {
		out[i] = 0 // 确保尾部补零
	}
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
