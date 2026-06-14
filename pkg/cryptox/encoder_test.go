package cryptox

import (
	"encoding/base64"
	"reflect"
	"testing"
)

func TestBase64Encoder(t *testing.T) {
	plaintext := []byte("hello")
	tests := []struct {
		name    string
		encoder encoder
		want    string
	}{
		{
			name:    "default",
			encoder: Base64Encoder{},
			want:    base64.StdEncoding.EncodeToString(plaintext),
		},
		{
			name:    "raw std",
			encoder: NewBase64Encoder(base64.RawStdEncoding),
			want:    base64.RawStdEncoding.EncodeToString(plaintext),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.encoder.EncodeToString(plaintext)
			if got != tt.want {
				t.Fatalf("EncodeToString() = %q, want %q", got, tt.want)
			}
			decoded, err := tt.encoder.DecodeString(got)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded, plaintext) {
				t.Fatalf("DecodeString() = %q, want %q", decoded, plaintext)
			}
		})
	}
}

func TestHexEncoder(t *testing.T) {
	plaintext := []byte("hello")
	encoder := HexEncoder{}
	got := encoder.EncodeToString(plaintext)
	if got != "68656c6c6f" {
		t.Fatalf("EncodeToString() = %q, want %q", got, "68656c6c6f")
	}
	decoded, err := encoder.DecodeString(got)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, plaintext) {
		t.Fatalf("DecodeString() = %q, want %q", decoded, plaintext)
	}
}
