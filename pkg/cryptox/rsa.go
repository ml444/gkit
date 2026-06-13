package cryptox

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"hash"

	"github.com/ml444/gkit/log"
)

type RSA struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	encoder    encoder
	ihash      hash.Hash
}

func NewRSA(private []byte) (*RSA, error) {
	r := &RSA{
		encoder: base64.StdEncoding,
		ihash:   sha256.New(),
	}
	err := r.SetPrivateKey(private)
	if err != nil {
		log.Errorf("err: %v\n", err)
		return nil, err
	}
	return r, nil
}

func (r *RSA) Encrypt(plaintext any) (any, error) {
	switch v := plaintext.(type) {
	case string:
		return r.EncryptWithString(v)
	case []byte:
		return r.EncryptWithBytes(v)
	default:
		return nil, errors.New("unsupported type")
	}
}

func (r *RSA) Decrypt(ciphertext any) (any, error) {
	switch v := ciphertext.(type) {
	case string:
		return r.DecryptWithString(v)
	case []byte:
		return r.DecryptWithBytes(v)
	default:
		return nil, errors.New("unsupported type")
	}
}

// EncryptWithBytes encrypts AdditionalData with public key
func (r *RSA) EncryptWithBytes(msg []byte) ([]byte, error) {
	// hash := sha512.New()
	return rsa.EncryptOAEP(r.ihash, rand.Reader, r.publicKey, msg, nil)
}

func (r *RSA) EncryptWithString(plaintext string) (string, error) {
	b, err := r.EncryptWithBytes([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return r.encoder.EncodeToString(b), nil
}

// DecryptWithBytes decrypts AdditionalData with private key
func (r *RSA) DecryptWithBytes(ciphertext []byte) ([]byte, error) {
	// hash := sha512.New()
	return rsa.DecryptOAEP(r.ihash, rand.Reader, r.privateKey, ciphertext, nil)
}

func (r *RSA) DecryptWithString(ciphertext string) (string, error) {
	cipherBuf, err := r.encoder.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	b, err := r.DecryptWithBytes(cipherBuf)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// EncryptPKCS1v15 encrypts AdditionalData with public key
func (r *RSA) EncryptPKCS1v15(msg []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, r.publicKey, msg)
}

// DecryptPKCS1v15 decrypts AdditionalData with private key
func (r *RSA) DecryptPKCS1v15(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, r.privateKey, ciphertext)
}

// SetPublicKey bytes to public key
func (r *RSA) SetPublicKey(public []byte) error {
	block, _ := pem.Decode(public)
	if block == nil {
		return errors.New("public key error")
	}
	ifc, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		return errors.New("ifc.(*rsa.PublicKey) not ok")
	}
	r.publicKey = key
	return nil
}

// SetPrivateKey bytes to private key
func (r *RSA) SetPrivateKey(private []byte) error {
	block, _ := pem.Decode(private)
	if block == nil {
		return errors.New("private key error")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return err
	}
	priv, ok := key.(*rsa.PrivateKey)
	if !ok {
		return errors.New("cryptox: parsed key is not an RSA private key")
	}
	r.privateKey = priv
	r.publicKey = &r.privateKey.PublicKey
	return nil
}

func ParsePrivatePem(privatePEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privatePEM)
	if block == nil {
		return nil, errors.New("decode private PEM error")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	priv, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("cryptox: parsed key is not an RSA private key")
	}
	return priv, nil
}

// SetEncoder sets the encoder for encoding and decoding
func (r *RSA) SetEncoder(encoder encoder) {
	r.encoder = encoder
}

// SetHash sets the hash function for signing and verifying
func (r *RSA) SetHash(h hash.Hash) {
	r.ihash = h
}

func GenerateRSAKey(bits int) (privateBytes, publicBytes []byte, err error) {
	if bits <= 0 {
		bits = 2048
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	privateBuf, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return
	}
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateBuf,
	}
	privateBytes = pem.EncodeToMemory(block)

	publicKey := &privateKey.PublicKey
	publicBuf, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return
	}
	block = &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicBuf,
	}
	publicBytes = pem.EncodeToMemory(block)
	return
}

func FormatPublicPEM(publicKey any) ([]byte, error) {
	publicBuf, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicBuf,
	}), nil
}

func FormatPrivatePEM(privateKey any) ([]byte, error) {
	privateBuf, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateBuf,
	}), nil
}
