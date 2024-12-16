package cryptox

import "encoding/hex"

var _ encoder = &HexEncoder{}

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
