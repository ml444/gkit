package crypto

import (
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/ml444/gkit/pkg/cryptox"
)

type Primitive struct {
	mu  sync.Mutex
	aes map[string]*cryptox.AES
	siv map[string]*cryptox.AESSIV
	enc *base64.Encoding
}

func NewPrimitive() *Primitive {
	return &Primitive{
		aes: make(map[string]*cryptox.AES),
		siv: make(map[string]*cryptox.AESSIV),
		enc: base64.RawStdEncoding,
	}
}

func (p *Primitive) Encrypt(mode Mode, key Key, plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	var raw []byte
	var err error
	switch mode {
	case ModeStorage:
		var c *cryptox.AES
		c, err = p.aesFor(key)
		if err != nil {
			return "", err
		}
		raw, err = c.EncryptWithBytes([]byte(plaintext))
	case ModeSearchable:
		var c *cryptox.AESSIV
		c, err = p.sivFor(key)
		if err != nil {
			return "", err
		}
		raw, err = c.EncryptWithBytes([]byte(plaintext))
	default:
		return "", fmt.Errorf("%w: invalid mode", ErrInvalidConfig)
	}
	if err != nil {
		return "", err
	}
	return FormatEnvelope(Envelope{
		Version: EnvelopeVersion,
		KeyID:   key.ID,
		Payload: p.enc.EncodeToString(raw),
	}), nil
}

func (p *Primitive) Decrypt(mode Mode, ring KeyRing, ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	env, err := ParseEnvelope(ciphertext)
	if err != nil {
		return "", err
	}
	key, err := ring.Get(env.KeyID)
	if err != nil {
		return "", err
	}
	raw, err := p.enc.DecodeString(env.Payload)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptFailed, err)
	}
	var plain []byte
	switch mode {
	case ModeStorage:
		var c *cryptox.AES
		c, err = p.aesFor(key)
		if err != nil {
			return "", err
		}
		plain, err = c.DecryptWithBytes(raw)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrDecryptFailed, err)
		}
	case ModeSearchable:
		var c *cryptox.AESSIV
		c, err = p.sivFor(key)
		if err != nil {
			return "", err
		}
		plain, err = c.DecryptWithBytes(raw)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrDecryptFailed, err)
		}
	default:
		return "", fmt.Errorf("%w: invalid mode", ErrInvalidConfig)
	}
	return string(plain), nil
}

func (p *Primitive) aesFor(key Key) (*cryptox.AES, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if c, ok := p.aes[key.ID]; ok {
		return c, nil
	}
	c, err := cryptox.NewAES(key.Material)
	if err != nil {
		return nil, fmt.Errorf("%w: aes: %v", ErrInvalidConfig, err)
	}
	p.aes[key.ID] = c
	return c, nil
}

func (p *Primitive) sivFor(key Key) (*cryptox.AESSIV, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if c, ok := p.siv[key.ID]; ok {
		return c, nil
	}
	c, err := cryptox.NewAESSIV(key.Material)
	if err != nil {
		return nil, fmt.Errorf("%w: siv: %v", ErrInvalidConfig, err)
	}
	p.siv[key.ID] = c
	return c, nil
}
