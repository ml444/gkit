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
	Bind(interface{}) error
	BindVars(interface{}) error
	BindQuery(interface{}) error
	BindForm(interface{}) error
	Returns(interface{}, error) error
	Result(int, interface{}) error
	JSON(int, interface{}) error
	XML(int, interface{}) error
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

func NewCtx(rsp http.ResponseWriter, req *http.Request) Context {
	return &wrappedCtx{
		status: http.StatusOK,
		coder:  &routerCoder{},
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
func (c *wrappedCtx) Bind(v interface{}) error      { return c.coder.BindBody()(c.req, v) }
func (c *wrappedCtx) BindVars(v interface{}) error  { return c.coder.BindVars()(c.req, v) }
func (c *wrappedCtx) BindQuery(v interface{}) error { return c.coder.BindQuery()(c.req, v) }
func (c *wrappedCtx) BindForm(v interface{}) error  { return c.coder.BindForm()(c.req, v) }

func (c *wrappedCtx) Returns(v interface{}, err error) error {
	if err != nil {
		return err
	}
	return c.Result(c.status, v)
}

func (c *wrappedCtx) Result(status int, v interface{}) error {
	return c.coder.ResponseEncoder()(status, c.rsp, c.req, v)
}

func (c *wrappedCtx) JSON(status int, v interface{}) error {
	c.rsp.Header().Set("Content-Type", "application/json")
	c.rsp.WriteHeader(status)
	return json.NewEncoder(c.rsp).Encode(v)
}

func (c *wrappedCtx) XML(status int, v interface{}) error {
	c.rsp.Header().Set("Content-Type", "application/xml")
	c.rsp.WriteHeader(status)
	return xml.NewEncoder(c.rsp).Encode(v)
}

func (c *wrappedCtx) String(status int, text string) error {
	c.rsp.Header().Set("Content-Type", "text/plain")
	c.rsp.WriteHeader(status)
	_, err := c.rsp.Write([]byte(text))
	if err != nil {
		return err
	}
	return nil
}

func (c *wrappedCtx) Blob(status int, contentType string, data []byte) error {
	c.rsp.Header().Set("Content-Type", contentType)
	c.rsp.WriteHeader(status)
	_, err := c.rsp.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (c *wrappedCtx) Stream(status int, contentType string, rd io.Reader) error {
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

func (c *wrappedCtx) Value(key interface{}) interface{} {
	if c.req == nil {
		return nil
	}
	return c.req.Context().Value(key)
}
