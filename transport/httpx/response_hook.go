package httpx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// ResponseMeta describes the final HTTP response before decoding. Header is an
// independent copy for each hook; ContentLength may be -1 and is not a size guarantee.
type ResponseMeta struct {
	StatusCode    int
	Header        http.Header
	ContentLength int64
}

type ResponseHeadersHook func(context.Context, ResponseMeta) error

type ResponseHookError struct {
	StatusCode int
	Err        error
}

func (e *ResponseHookError) Error() string {
	return fmt.Sprintf("httpx: response headers hook (HTTP %d): %v", e.StatusCode, e.Err)
}
func (e *ResponseHookError) Unwrap() error { return e.Err }

// WithResponseHeadersHook appends a hook for managed calls (Invoke, InvokeReader,
// Download). Hooks run in registration order, before per-call hooks and decoding.
// Raw Do retains its net/http semantics and does not execute hooks.
func WithResponseHeadersHook(h ResponseHeadersHook) ClientOption {
	return func(c *Client) {
		if h == nil {
			c.configErr = errors.New("httpx: nil response headers hook")
			return
		}
		c.responseHooks = append(c.responseHooks, h)
	}
}

// OnResponseHeaders appends a per-call hook. An error stops remaining hooks and
// decoding; the call closes the body and returns a ResponseHookError.
func OnResponseHeaders(h ResponseHeadersHook) CallOption {
	return func(c *callInfo) {
		if h == nil {
			c.configErr = errors.New("httpx: nil response headers hook")
			return
		}
		c.responseHooks = append(c.responseHooks, h)
	}
}

func runResponseHeadersHooks(ctx context.Context, r *http.Response, groups ...[]ResponseHeadersHook) error {
	for _, hooks := range groups {
		for _, hook := range hooks {
			header := r.Header.Clone()
			if header == nil {
				header = make(http.Header)
			}
			meta := ResponseMeta{StatusCode: r.StatusCode, Header: header, ContentLength: r.ContentLength}
			if err := hook(ctx, meta); err != nil {
				return &ResponseHookError{StatusCode: r.StatusCode, Err: err}
			}
		}
	}
	return nil
}
