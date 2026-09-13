package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/transport"
)

func parseCallOptions(path string, opts []CallOption) (callInfo, error) {
	c := defaultCallInfo(path)
	for _, opt := range opts {
		if opt == nil {
			return c, errors.New("httpx: nil call option")
		}
		opt(&c)
	}
	return c, c.configErr
}

func (c *Client) validateInvocation(path string) error {
	if c.target == nil || c.target.Authority == "" {
		return errors.New("httpx: invocation requires an endpoint or discovery service")
	}
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return errors.New("httpx: invocation path must start with a single slash")
	}
	return nil
}

func (c *Client) prepareInvocation(ctx context.Context, method, path string, body io.Reader, call callInfo) (context.Context, *http.Request, error) {
	if err := c.validateInvocation(path); err != nil {
		return ctx, nil, err
	}
	u := fmt.Sprintf("%s://%s%s", c.target.Scheme, c.target.Authority, path)
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return ctx, nil, err
	}
	if call.reqHeader != nil {
		req.Header = call.reqHeader.Clone()
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	ctx = transport.ToContext(ctx, &Transport{
		endpoint: c.endpoint, path: call.operation, inMD: transport.New(req.Header),
		pathTemplate: call.pathTemplate, req: req,
	})
	return ctx, req, nil
}

type responseHandler struct {
	received func(*http.Response)
	consume  func(context.Context, *http.Response) (interface{}, error)
}

// execute owns bodies across all exits, including middleware short circuits.
// Stream methods permit only one downstream invocation: replay could duplicate
// writes to dst or consume an already exhausted upload reader.
func (c *Client) execute(ctx context.Context, req *http.Request, args interface{}, call callInfo, handler responseHandler, singleUse bool) (interface{}, error) {
	if req.Body != nil {
		req.Body = closeOnce(req.Body)
		defer req.Body.Close()
	}
	var used atomic.Bool
	h := func(ctx context.Context, _ interface{}) (out interface{}, err error) {
		if singleUse && !used.CompareAndSwap(false, true) {
			return nil, errors.New("httpx: streaming calls cannot be replayed by middleware")
		}
		holder := &instanceHolder{}
		ctx = context.WithValue(ctx, instanceHolderKey{}, holder)
		outReq := req.Clone(ctx)
		if tr, ok := transport.FromContext(ctx); ok {
			outReq.Header = make(http.Header)
			for key, values := range tr.In() {
				for _, value := range values {
					outReq.Header.Add(key, value)
				}
			}
		}
		res, err := c.Do(outReq)
		if err != nil {
			if ctx.Err() == nil {
				c.updateDiscoveryStatus(ctx, holder, false)
			}
			return nil, err
		}
		body := &observedResponseBody{ReadCloser: closeOnce(res.Body)}
		res.Body = body
		defer res.Body.Close()
		remoteError := false
		defer func() {
			if success, report := responseFeedback(ctx, res.StatusCode, body.readErr, err, remoteError); report {
				c.updateDiscoveryStatus(ctx, holder, success)
			}
		}()
		if handler.received != nil {
			handler.received(res)
		}
		if err = runResponseHeadersHooks(ctx, res, c.responseHooks, call.responseHooks); err != nil {
			return nil, err
		}
		out, err = handler.consume(ctx, res)
		if err != nil {
			var statusErr *UnexpectedStatusError
			var businessErr *errorx.Error
			remoteError = (errors.As(err, &businessErr) && int(businessErr.Status) == res.StatusCode) || errors.As(err, &statusErr)
			return out, err
		}
		if call.onResponse != nil {
			if err = call.onResponse(res); err != nil {
				return nil, err
			}
		}
		return out, err
	}
	if len(c.middleware) > 0 {
		h = middleware.Chain(c.middleware...)(h)
	}
	return h(ctx, args)
}

func responseFeedback(ctx context.Context, status int, readErr, callErr error, remoteError bool) (success, report bool) {
	if status >= 500 {
		return false, true
	}
	if ctx.Err() != nil {
		return false, false
	}
	if readErr != nil {
		return false, true
	}
	if callErr == nil || remoteError {
		return true, true
	}
	return false, false // local hook/decoder/writer errors do not penalize a node
}

type onceReadCloser struct {
	io.ReadCloser
	once sync.Once
	err  error
}

func closeOnce(r io.ReadCloser) *onceReadCloser {
	if once, ok := r.(*onceReadCloser); ok {
		return once
	}
	return &onceReadCloser{ReadCloser: r}
}

func (r *onceReadCloser) Close() error {
	r.once.Do(func() { r.err = r.ReadCloser.Close() })
	return r.err
}

type observedResponseBody struct {
	io.ReadCloser
	readErr error
}

func (r *observedResponseBody) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if err != nil && !errors.Is(err, io.EOF) {
		r.readErr = err
	}
	return n, err
}
