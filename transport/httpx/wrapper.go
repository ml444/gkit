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

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/transport/httpx/encoding/form"
	jsonx "github.com/ml444/gkit/transport/httpx/encoding/json"
)

var _ Context = (*wrapper)(nil)
var _ = jsonx.Name

func defaultError(err error) error {
	return errorx.CreateError(errorx.DefaultStatusCode, errorx.ErrCodeInvalidReqSys, err.Error())
}

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
}

type responseWriter struct {
	code int
	w    http.ResponseWriter
}

func (w *responseWriter) reset(res http.ResponseWriter) {
	w.w = res
	w.code = http.StatusOK
}
func (w *responseWriter) Header() http.Header        { return w.w.Header() }
func (w *responseWriter) WriteHeader(statusCode int) { w.code = statusCode }
func (w *responseWriter) Write(data []byte) (int, error) {
	w.w.WriteHeader(w.code)
	return w.w.Write(data)
}

type wrapper struct {
	router *Router
	req    *http.Request
	res    http.ResponseWriter
	w      responseWriter
}

func (c *wrapper) Header() http.Header {
	return c.req.Header
}

func (c *wrapper) Vars() url.Values {
	raws := mux.Vars(c.req)
	vars := make(url.Values, len(raws))
	for k, v := range raws {
		vars[k] = []string{v}
	}
	return vars
}

func (c *wrapper) Form() url.Values {
	if err := c.req.ParseForm(); err != nil {
		return url.Values{}
	}
	return c.req.Form
}

func (c *wrapper) Query() url.Values {
	return c.req.URL.Query()
}
func (c *wrapper) Request() *http.Request        { return c.req }
func (c *wrapper) Response() http.ResponseWriter { return c.res }
func (c *wrapper) Bind(v interface{}) error      { return DefaultRequestDecoder(c.req, v) }
func (c *wrapper) BindVars(v interface{}) error  { return DefaultRequestVars(c.req, v) }
func (c *wrapper) BindQuery(v interface{}) error {
	if err := getCoder(form.Name).Unmarshal([]byte(c.req.URL.Query().Encode()), v); err != nil {
		return defaultError(err)
	}
	return nil
}
func (c *wrapper) BindForm(v interface{}) error {
	if err := c.req.ParseForm(); err != nil {
		return err
	}
	if err := getCoder(form.Name).Unmarshal([]byte(c.req.Form.Encode()), v); err != nil {
		return defaultError(err)
	}
	return nil
}

func (c *wrapper) Returns(v interface{}, err error) error {
	if err != nil {
		return err
	}
	return DefaultResponseEncoder(&c.w, c.req, v)
}

func (c *wrapper) Result(code int, v interface{}) error {
	c.w.WriteHeader(code)
	return DefaultResponseEncoder(&c.w, c.req, v)
}

func (c *wrapper) JSON(code int, v interface{}) error {
	c.res.Header().Set("Content-Type", "application/json")
	c.res.WriteHeader(code)
	return json.NewEncoder(c.res).Encode(v)
}

func (c *wrapper) XML(code int, v interface{}) error {
	c.res.Header().Set("Content-Type", "application/xml")
	c.res.WriteHeader(code)
	return xml.NewEncoder(c.res).Encode(v)
}

func (c *wrapper) String(code int, text string) error {
	c.res.Header().Set("Content-Type", "text/plain")
	c.res.WriteHeader(code)
	_, err := c.res.Write([]byte(text))
	if err != nil {
		return err
	}
	return nil
}

func (c *wrapper) Blob(code int, contentType string, data []byte) error {
	c.res.Header().Set("Content-Type", contentType)
	c.res.WriteHeader(code)
	_, err := c.res.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (c *wrapper) Stream(code int, contentType string, rd io.Reader) error {
	c.res.Header().Set("Content-Type", contentType)
	c.res.WriteHeader(code)
	_, err := io.Copy(c.res, rd)
	return err
}

func (c *wrapper) Reset(res http.ResponseWriter, req *http.Request) {
	c.w.reset(res)
	c.res = res
	c.req = req
}

func (c *wrapper) Deadline() (time.Time, bool) {
	if c.req == nil {
		return time.Time{}, false
	}
	return c.req.Context().Deadline()
}

func (c *wrapper) Done() <-chan struct{} {
	if c.req == nil {
		return nil
	}
	return c.req.Context().Done()
}

func (c *wrapper) Err() error {
	if c.req == nil {
		return context.Canceled
	}
	return c.req.Context().Err()
}

func (c *wrapper) Value(key interface{}) interface{} {
	if c.req == nil {
		return nil
	}
	return c.req.Context().Value(key)
}
