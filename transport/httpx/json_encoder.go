package httpx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// JSONMode selects the encoding used by Context.JSON.
type JSONMode uint8

const (
	JSONLegacy JSONMode = iota // Standard library Encoder, including its trailing newline.
	JSONCodec                  // This router coder's configured JSON codec.
)

// JSONEncoderProvider optionally extends IRouterCoder without changing its contract.
type JSONEncoderProvider interface {
	JSONEncoder() ResponseEncoder
}

// WithJSONMode selects Context.JSON behavior. JSONLegacy is the default.
func WithJSONMode(mode JSONMode) RouterCoderOption {
	return func(c *routerCoder) error {
		if mode != JSONLegacy && mode != JSONCodec {
			return fmt.Errorf("httpx: invalid JSON mode %d", mode)
		}
		c.jsonMode = mode
		return nil
	}
}

func (c *routerCoder) JSONEncoder() ResponseEncoder {
	if c.jsonMode == JSONLegacy {
		return legacyJSONEncoder
	}
	return func(status int, w http.ResponseWriter, r *http.Request, v interface{}) error {
		if !responseAllowsBody(status, r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			return nil
		}
		data, err := c.profile.get("json").Marshal(v)
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		n, err := w.Write(data)
		if err == nil && n != len(data) {
			err = io.ErrShortWrite
		}
		if err != nil {
			return &responseWriteError{err: err}
		}
		return nil
	}
}

func responseAllowsBody(status int, r *http.Request) bool {
	return status >= 200 && status != 204 && status != 205 && status != 304 && (r == nil || r.Method != http.MethodHead)
}

func legacyJSONEncoder(status int, w http.ResponseWriter, _ *http.Request, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return &responseWriteError{err: err}
	}
	return nil
}
