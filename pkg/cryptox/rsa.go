package cryptox

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"io"

	"github.com/ml444/gkit/log"
)

type RSA struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	encoder    encoder
}

func NewRSA(private []byte) (*RSA, error) {
	r := &RSA{}
	err := r.SetPrivateKey(private)
	if err != nil {
		log.Errorf("err: %v\n", err)
		return nil, err
	}

	if r.encoder == nil {
		r.encoder = base64.RawStdEncoding
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
	hash := sha512.New()
	return rsa.EncryptOAEP(hash, rand.Reader, r.publicKey, msg, nil)
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
	hash := sha512.New()
	return rsa.DecryptOAEP(hash, rand.Reader, r.privateKey, ciphertext, nil)
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
		return errors.New("private key error")
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
	r.privateKey = key.(*rsa.PrivateKey)
	r.publicKey = &r.privateKey.PublicKey
	return nil
}

// SetPublicKeyByBase64 Get bytes AdditionalData by decoding base64 string
func (r *RSA) SetPublicKeyByBase64(publicStr string) error {
	return r.SetPublicKey([]byte(publicStr))
}

// SetPrivateKeyByBase64 Get bytes AdditionalData by decoding base64 string
func (r *RSA) SetPrivateKeyByBase64(privateStr string) error {
	return r.SetPrivateKey([]byte(privateStr))
}

func GenRSAKey(out io.Writer, bits int) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	privateStream := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateStream,
	}
	prvKeyBuf := pem.EncodeToMemory(block)
	n, err := out.Write(prvKeyBuf)
	if err != nil {
		if err == io.ErrShortWrite {
			for n < len(prvKeyBuf) {
				x, err := out.Write(prvKeyBuf[n:])
				if err != nil {
					return err
				}
				n += x
			}
		} else {
			return err
		}
		return err
	}
	return nil
}
