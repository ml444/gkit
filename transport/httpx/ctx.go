package httpx

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"

	"github.com/ml444/gkit/transport"
)

var _ Context = (*wrappedCtx)(nil)

// Context is an HTTP Context.
type Context interface {
	context.Context
	Vars() url.Values
	Query() url.Values
	Form() url.Values
	Header() http.Header
	Request() *http.Request
	Response() http.ResponseWriter
	Bind(any) error
	BindVars(any) error
	BindQuery(any) error
	BindForm(any) error
	Returns(any, error)
	Result(int, any)
	JSON(int, any) error
	XML(int, any) error
	String(int, string) error
	Blob(int, string, []byte) error
	Stream(int, string, io.Reader) error
	Reset(http.ResponseWriter, *http.Request)
	ReturnError(error)
}

type wrappedCtx struct {
	status int
	coder  IRouterCoder
	req    *http.Request
	rsp    http.ResponseWriter
}

type routerCoderKey struct{}

func NewCtx(rsp http.ResponseWriter, req *http.Request) Context {
	var rc IRouterCoder
	if req != nil {
		if v := req.Context().Value(routerCoderKey{}); v != nil {
			if c, ok := v.(IRouterCoder); ok && c != nil {
				rc = c
			}
		}
	}
	if rc == nil {
		rc = newRouterCoder()
	}
	return &wrappedCtx{
		status: http.StatusOK,
		coder:  rc,
		req:    req,
		rsp:    rsp,
	}
}

func (c *wrappedCtx) Header() http.Header {
	return c.req.Header
}

func (c *wrappedCtx) Vars() url.Values {
	raws := mux.Vars(c.req)
	vars := make(url.Values, len(raws))
	for k, v := range raws {
		vars[k] = []string{v}
	}
	return vars
}

func (c *wrappedCtx) Form() url.Values {
	if err := c.req.ParseForm(); err != nil {
		return url.Values{}
	}
	return c.req.Form
}

func (c *wrappedCtx) Query() url.Values {
	return c.req.URL.Query()
}
func (c *wrappedCtx) Request() *http.Request        { return c.req }
func (c *wrappedCtx) Response() http.ResponseWriter { return c.rsp }
func (c *wrappedCtx) Bind(v any) error              { return c.coder.BindBody()(c.req, v) }
func (c *wrappedCtx) BindVars(v any) error          { return c.coder.BindVars()(c.req, v) }
func (c *wrappedCtx) BindQuery(v any) error         { return c.coder.BindQuery()(c.req, v) }
func (c *wrappedCtx) BindForm(v any) error          { return c.coder.BindForm()(c.req, v) }

func (c *wrappedCtx) Returns(v any, err error) {
	if err != nil {
		c.ReturnError(err)
		return
	}
	c.Result(c.status, v)
}

func (c *wrappedCtx) Result(status int, v any) {
	c.setResponseHeaders()
	err := c.coder.ResponseEncoder()(status, c.rsp, c.req, v)
	if err != nil {
		c.ReturnError(err)
		return
	}
}

func (c *wrappedCtx) JSON(status int, v any) error {
	c.setResponseHeaders()
	c.rsp.Header().Set("Content-Type", "application/json")
	c.rsp.WriteHeader(status)
	return json.NewEncoder(c.rsp).Encode(v)
}

func (c *wrappedCtx) XML(status int, v any) error {
	c.setResponseHeaders()
	c.rsp.Header().Set("Content-Type", "application/xml")
	c.rsp.WriteHeader(status)
	return xml.NewEncoder(c.rsp).Encode(v)
}

func (c *wrappedCtx) String(status int, text string) error {
	c.setResponseHeaders()
	c.rsp.Header().Set("Content-Type", "text/plain")
	c.rsp.WriteHeader(status)
	_, err := c.rsp.Write([]byte(text))
	if err != nil {
		return err
	}
	return nil
}

func (c *wrappedCtx) Blob(status int, contentType string, data []byte) error {
	c.setResponseHeaders()
	c.rsp.Header().Set("Content-Type", contentType)
	c.rsp.WriteHeader(status)
	_, err := c.rsp.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (c *wrappedCtx) Stream(status int, contentType string, rd io.Reader) error {
	c.setResponseHeaders()
	c.rsp.Header().Set("Content-Type", contentType)
	c.rsp.WriteHeader(status)
	_, err := io.Copy(c.rsp, rd)
	return err
}

func (c *wrappedCtx) Reset(rsp http.ResponseWriter, req *http.Request) {
	c.rsp = rsp
	c.req = req
}

func (c *wrappedCtx) ReturnError(err error) {
	c.setResponseHeaders()
	c.coder.ErrorEncoder()(c.rsp, c.req, err)
}

func (c *wrappedCtx) Deadline() (time.Time, bool) {
	if c.req == nil {
		return time.Time{}, false
	}
	return c.req.Context().Deadline()
}

func (c *wrappedCtx) Done() <-chan struct{} {
	if c.req == nil {
		return nil
	}
	return c.req.Context().Done()
}

func (c *wrappedCtx) Err() error {
	if c.req == nil {
		return context.Canceled
	}
	return c.req.Context().Err()
}

func (c *wrappedCtx) Value(key any) any {
	if c.req == nil {
		return nil
	}
	return c.req.Context().Value(key)
}

func (c *wrappedCtx) setResponseHeaders() {
	if c.req == nil {
		return
	}
	ctx := c.req.Context()
	tr, ok := transport.FromContext(ctx)
	if !ok {
		return
	}
	if outHeaders := tr.Out(); len(outHeaders) > 0 {
		for k, v := range outHeaders {
			if len(v) == 0 {
				continue
			}
			c.rsp.Header().Add(k, v[0])
		}
	}
}
