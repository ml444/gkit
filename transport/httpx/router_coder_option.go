package httpx

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/ml444/gkit/transport/httpx/coder"
)

// RouterCoderOption configures a router coder before use.
type RouterCoderOption func(*routerCoder) error

// NewRouterCoder snapshots the registered codecs and the built-in JSON options,
// then applies opts in order. Later options override earlier ones. Omitted
// callbacks retain their defaults; explicit nil values are configuration errors.
// Callbacks and codec objects must be safe for concurrent use.
func NewRouterCoder(opts ...RouterCoderOption) (IRouterCoder, error) {
	c := newRouterCoderWithProfile(defaultCodecProfile())
	for _, opt := range opts {
		if opt == nil {
			return nil, errors.New("httpx: nil router coder option")
		}
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

// WithBaseRouterCoder replaces all six callbacks. Put it before individual
// overrides to retain those overrides. The base's own codec configuration is
// retained by its callbacks; WithJSONCoder only affects this factory's defaults.
func WithBaseRouterCoder(base IRouterCoder) RouterCoderOption {
	return func(c *routerCoder) error {
		if isNilCoderOptionValue(base) {
			return errors.New("httpx: base router coder is required")
		}
		c.bindVars, c.bindQuery = base.BindVars(), base.BindQuery()
		c.bindForm, c.bindBody = base.BindForm(), base.BindBody()
		c.respEnc, c.errEnc = base.ResponseEncoder(), base.ErrorEncoder()
		return c.validate()
	}
}

// WithBindVars replaces the path parameter decoder.
func WithBindVars(fn RequestDecoder) RouterCoderOption {
	return func(c *routerCoder) error {
		if fn == nil {
			return errors.New("httpx: path decoder is required")
		}
		c.bindVars = fn
		return nil
	}
}

// WithBindQuery replaces the query parameter decoder.
func WithBindQuery(fn RequestDecoder) RouterCoderOption {
	return func(c *routerCoder) error {
		if fn == nil {
			return errors.New("httpx: query decoder is required")
		}
		c.bindQuery = fn
		return nil
	}
}

// WithBindForm replaces the form decoder.
func WithBindForm(fn RequestDecoder) RouterCoderOption {
	return func(c *routerCoder) error {
		if fn == nil {
			return errors.New("httpx: form decoder is required")
		}
		c.bindForm = fn
		return nil
	}
}

// WithBindBody replaces the request body decoder. A custom decoder is responsible
// for its own codec policy and should preserve request size-limit errors.
func WithBindBody(fn RequestDecoder) RouterCoderOption {
	return func(c *routerCoder) error {
		if fn == nil {
			return errors.New("httpx: body decoder is required")
		}
		c.bindBody = fn
		return nil
	}
}

// WithResponseEncoder replaces the normal response encoder.
func WithResponseEncoder(fn ResponseEncoder) RouterCoderOption {
	return func(c *routerCoder) error {
		if fn == nil {
			return errors.New("httpx: response encoder is required")
		}
		c.respEnc = fn
		return nil
	}
}

// WithErrorEncoder replaces the error response encoder.
func WithErrorEncoder(fn ErrorEncoder) RouterCoderOption {
	return func(c *routerCoder) error {
		if fn == nil {
			return errors.New("httpx: error encoder is required")
		}
		c.errEnc = fn
		return nil
	}
}

// WithJSONCoder replaces this instance's JSON codec for default body binding,
// normal responses, error responses and JSON fallback. It does not change
// Context.JSON, client codecs, or callbacks supplied via other options.
func WithJSONCoder(codec coder.ICoder) RouterCoderOption {
	return func(c *routerCoder) error {
		if isNilCoderOptionValue(codec) || codec.Name() != "json" {
			return errors.New("httpx: JSON codec must be non-nil and named json")
		}
		c.profile.codecs["json"] = codec
		return nil
	}
}

func (c *routerCoder) validate() error {
	for _, callback := range []struct {
		name string
		fn   interface{}
	}{
		{"BindVars", c.bindVars}, {"BindQuery", c.bindQuery},
		{"BindForm", c.bindForm}, {"BindBody", c.bindBody},
		{"ResponseEncoder", c.respEnc}, {"ErrorEncoder", c.errEnc},
	} {
		if isNilCoderOptionValue(callback.fn) {
			return fmt.Errorf("httpx: %s callback is required", callback.name)
		}
	}
	return nil
}

func isNilCoderOptionValue(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
