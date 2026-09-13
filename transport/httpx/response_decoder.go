package httpx

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/transport/httpx/coder"
)

// ResponseTooLargeError retains the remote status and the configured byte limit.
// Unwrap preserves the legacy errorx 502/50201 classification.
type ResponseTooLargeError struct {
	Limit      int64
	StatusCode int
}

func (*ResponseTooLargeError) Error() string { return "httpx: response body too large" }
func (*ResponseTooLargeError) Unwrap() error {
	return errorx.CreateError(502, 50201, "response body too large")
}

type responseDecodeConfig struct {
	limit   int64
	profile *codecProfile
}

type ResponseDecoderOption func(*responseDecodeConfig) error

// NewResponseDecoder snapshots codec configuration. The default limit is 10 MiB.
// Use the result with WithResponseDecoder. It does not affect raw Do responses.
func NewResponseDecoder(opts ...ResponseDecoderOption) (DecodeResponseFunc, error) {
	c := responseDecodeConfig{limit: maxRespBytes, profile: defaultCodecProfile()}
	for _, opt := range opts {
		if opt == nil {
			return nil, errors.New("httpx: nil response decoder option")
		}
		if err := opt(&c); err != nil {
			return nil, err
		}
	}
	return func(_ context.Context, r *http.Response, out interface{}) error {
		return decodeResponse(r, out, c.limit, func(media string) coder.ICoder {
			return c.profile.get(contentSubtype(media))
		})
	}, nil
}

// WithDecodeMaxBytes limits bytes consumed by decoding, including error bodies.
// Zero is unlimited; negative values are invalid. Skipped success bodies are not read.
func WithDecodeMaxBytes(n int64) ResponseDecoderOption {
	return func(c *responseDecodeConfig) error {
		if n < 0 {
			return errors.New("httpx: decode limit must not be negative")
		}
		c.limit = n
		return nil
	}
}

// WithDecodeJSONCoder replaces JSON decoding and JSON fallback for this decoder.
func WithDecodeJSONCoder(codec coder.ICoder) ResponseDecoderOption {
	return func(c *responseDecodeConfig) error {
		if isNilCoderOptionValue(codec) || codec.Name() != "json" {
			return errors.New("httpx: JSON codec must be non-nil and named json")
		}
		c.profile.codecs["json"] = codec
		return nil
	}
}

func readResponseBytes(r *http.Response, limit int64) ([]byte, error) {
	var reader io.Reader = r.Body
	if limit > 0 {
		n := limit
		if n < math.MaxInt64 {
			n++
		}
		reader = io.LimitReader(reader, n)
	}
	data, err := io.ReadAll(reader)
	if limit > 0 && int64(len(data)) > limit {
		return nil, &ResponseTooLargeError{Limit: limit, StatusCode: r.StatusCode}
	}
	return data, err
}
