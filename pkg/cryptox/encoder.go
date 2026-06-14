package cryptox

import (
	"encoding/base64"
	"encoding/hex"
)

var (
	_ encoder = HexEncoder{}
	_ encoder = Base64Encoder{}
)

type encoder interface {
	EncodeToString([]byte) string
	DecodeString(string) ([]byte, error)
}

type HexEncoder struct{}

func (h HexEncoder) EncodeToString(b []byte) string {
	return hex.EncodeToString(b)
}

func (h HexEncoder) DecodeString(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

type Base64Encoder struct {
	Encoding *base64.Encoding
}

func NewBase64Encoder(encoding *base64.Encoding) Base64Encoder {
	if encoding == nil {
		encoding = base64.StdEncoding
	}
	return Base64Encoder{Encoding: encoding}
}

func (b Base64Encoder) EncodeToString(src []byte) string {
	return b.encoding().EncodeToString(src)
}

func (b Base64Encoder) DecodeString(s string) ([]byte, error) {
	return b.encoding().DecodeString(s)
}

func (b Base64Encoder) encoding() *base64.Encoding {
	if b.Encoding != nil {
		return b.Encoding
	}
	return base64.StdEncoding
}
