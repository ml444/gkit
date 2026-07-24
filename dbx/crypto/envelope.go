package crypto

import (
	"fmt"
	"strconv"
	"strings"
)

const EnvelopeVersion = 1

type Envelope struct {
	Version int
	KeyID   string
	Payload string
}

func FormatEnvelope(e Envelope) string {
	return fmt.Sprintf("v%d:k%s:%s", e.Version, e.KeyID, e.Payload)
}

func ParseEnvelope(s string) (Envelope, error) {
	// v{n}:k{id}:{payload} — payload may contain ':'
	if !strings.HasPrefix(s, "v") {
		return Envelope{}, fmt.Errorf("%w: missing version", ErrDecryptFailed)
	}
	rest := s[1:]
	i := strings.IndexByte(rest, ':')
	if i <= 0 {
		return Envelope{}, fmt.Errorf("%w: bad version", ErrDecryptFailed)
	}
	ver, err := strconv.Atoi(rest[:i])
	if err != nil || ver <= 0 {
		return Envelope{}, fmt.Errorf("%w: bad version", ErrDecryptFailed)
	}
	rest = rest[i+1:]
	if !strings.HasPrefix(rest, "k") {
		return Envelope{}, fmt.Errorf("%w: missing key id", ErrDecryptFailed)
	}
	rest = rest[1:]
	j := strings.IndexByte(rest, ':')
	if j <= 0 {
		return Envelope{}, fmt.Errorf("%w: bad key id", ErrDecryptFailed)
	}
	keyID := rest[:j]
	payload := rest[j+1:]
	if keyID == "" || payload == "" {
		return Envelope{}, fmt.Errorf("%w: empty key or payload", ErrDecryptFailed)
	}
	return Envelope{Version: ver, KeyID: keyID, Payload: payload}, nil
}
