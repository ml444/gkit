package httpx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/transport/httpx/coder"
	"github.com/ml444/gkit/transport/httpx/coder/form"
)

// responseWriteError marks a failure after the response status was committed.
// Callers must not try to encode a second HTTP response.
type responseWriteError struct{ err error }

func (e *responseWriteError) Error() string { return "httpx: response write: " + e.err.Error() }
func (e *responseWriteError) Unwrap() error { return e.err }

const maxRespBytes = 10 << 20 // 10MB，可配置

// RequestDecoder is decode request func.
type RequestDecoder func(*http.Request, interface{}) error

// ResponseEncoder is encode response func.
type ResponseEncoder func(int, http.ResponseWriter, *http.Request, interface{}) error

// ErrorEncoder is encode error func.
type ErrorEncoder func(http.ResponseWriter, *http.Request, error)

type IRouterCoder interface {
	BindVars() RequestDecoder
	BindQuery() RequestDecoder
	BindForm() RequestDecoder
	BindBody() RequestDecoder
	ResponseEncoder() ResponseEncoder
	ErrorEncoder() ErrorEncoder

	// SetBindVars(RequestDecoder)
	// SetBindQuery(RequestDecoder)
	// SetBindForm(RequestDecoder)
	// SetBindBody(RequestDecoder)
	// SetResponseEncoder(ResponseEncoder)
	// SetErrorEncoder(ErrorEncoder)
}

type routerCoder struct {
	bindVars  RequestDecoder
	bindQuery RequestDecoder
	bindForm  RequestDecoder
	bindBody  RequestDecoder
	respEnc   ResponseEncoder
	errEnc    ErrorEncoder
}

func newRouterCoder() *routerCoder {
	c := &routerCoder{}
	c.bindVars = func(r *http.Request, target interface{}) error {
		vars := requestVars(r)
		if err := coder.GetCoder(form.Name).Unmarshal([]byte(vars.Encode()), target); err != nil {
			return defaultError(err)
		}
		return nil
	}
	c.bindQuery = func(r *http.Request, v interface{}) error {
		if err := coder.GetCoder(form.Name).Unmarshal([]byte(r.URL.Query().Encode()), v); err != nil {
			return defaultError(err)
		}
		return nil
	}
	c.bindForm = func(r *http.Request, v interface{}) error {
		if err := r.ParseForm(); err != nil {
			return err
		}
		if err := coder.GetCoder(form.Name).Unmarshal([]byte(r.Form.Encode()), v); err != nil {
			return defaultError(err)
		}
		return nil
	}
	c.bindBody = func(r *http.Request, v interface{}) error {
		codec, _ := getCoderForRequest(r, "Content-Type")
		data, err := io.ReadAll(r.Body)

		// reset body.
		r.Body = io.NopCloser(bytes.NewBuffer(data))

		if err != nil {
			var mbe *http.MaxBytesError
			if errors.As(err, &mbe) {
				return errorx.CreateError(http.StatusRequestEntityTooLarge, errorx.ErrCodeInvalidReqSys, err.Error())
			}
			return errorx.BadRequest(err.Error())
		}
		if len(data) == 0 {
			return nil
		}
		if err = codec.Unmarshal(data, v); err != nil {
			return errorx.BadRequest(fmt.Sprintf("body unmarshal %s", err.Error()))
		}
		return nil
	}
	c.respEnc = func(status int, w http.ResponseWriter, r *http.Request, v interface{}) error {
		if v == nil {
			w.WriteHeader(status)
			return nil
		}
		if rd, ok := v.(IRedirect); ok {
			redirectUrl, code := rd.Redirect()
			http.Redirect(w, r, redirectUrl, code)
			return nil
		}
		codec, _ := getCoderForRequest(r, "Accept")
		data, err := codec.Marshal(v)
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", joinContentType(codec.Name()))
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
	c.errEnc = func(w http.ResponseWriter, r *http.Request, err error) {
		ex := errorx.FromError(err)
		ex.ConvertMsgByLang(getAcceptLanguage(r)...)
		codec, _ := getCoderForRequest(r, "Accept")
		body, err := codec.Marshal(ex)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", joinContentType(codec.Name()))
		w.WriteHeader(int(ex.Status))
		_, _ = w.Write(body)
	}
	return c
}

func (c *routerCoder) BindVars() RequestDecoder {
	return c.bindVars
}

func (c *routerCoder) BindQuery() RequestDecoder {
	return c.bindQuery
}

func (c *routerCoder) BindForm() RequestDecoder {
	return c.bindForm
}

func (c *routerCoder) BindBody() RequestDecoder {
	return c.bindBody
}

func (c *routerCoder) ResponseEncoder() ResponseEncoder {
	return c.respEnc
}

func (c *routerCoder) ErrorEncoder() ErrorEncoder {
	return c.errEnc
}

func DefaultRequestEncoder(_ context.Context, contentType string, in interface{}) ([]byte, error) {
	name := contentSubtype(contentType)
	body, err := coder.GetCoder(name).Marshal(in)
	if err != nil {
		return nil, err
	}
	return body, err
}

// DefaultResponseDecoder is an HTTP response decoder.
func DefaultResponseDecoder(_ context.Context, rsp *http.Response, v interface{}) error {
	if rsp.StatusCode < 400 && (v == nil || rsp.StatusCode == http.StatusNoContent || rsp.StatusCode == http.StatusResetContent || (rsp.Request != nil && rsp.Request.Method == http.MethodHead)) {
		return nil
	}
	data, err := io.ReadAll(io.LimitReader(rsp.Body, maxRespBytes+1))
	if err != nil {
		return err
	}
	if int64(len(data)) > maxRespBytes {
		return errorx.CreateError(502, 50201, "response body too large")
	}
	if rsp.StatusCode >= 400 {
		e := new(errorx.Error)
		if err = coderByContentType(rsp.Header.Get("Content-Type")).Unmarshal(data, e); err == nil {
			e.Status = int32(rsp.StatusCode)
			return e
		} else {
			e.Status = int32(rsp.StatusCode)
			e.Code = errorx.ErrCodeInvalidReqSys
			e.Message = string(data)
			return e
		}
	}
	return coderByContentType(rsp.Header.Get("Content-Type")).Unmarshal(data, v)
}

func getCoderForRequest(r *http.Request, name string) (coder.ICoder, bool) {
	for _, accept := range r.Header[name] {
		codec := coder.GetCoder(contentSubtype(accept))
		if codec != nil {
			return codec, true
		}
	}
	return coder.GetCoder("json"), false
}

func coderByContentType(contentType string) coder.ICoder {
	codec := coder.GetCoder(contentSubtype(contentType))
	if codec != nil {
		return codec
	}
	return coder.GetCoder("json")
}

func contentSubtype(contentType string) string {
	left := strings.Index(contentType, "/")
	if left == -1 {
		return ""
	}
	right := strings.Index(contentType, ";")
	if right == -1 {
		right = len(contentType)
	}
	if right < left {
		return ""
	}
	return contentType[left+1 : right]
}

const contentTypePrefix = "application"

func joinContentType(subtype string) string {
	return strings.Join([]string{contentTypePrefix, subtype}, "/")
}

func defaultError(err error) error {
	return errorx.CreateError(errorx.DefaultStatusCode, errorx.ErrCodeInvalidReqSys, err.Error())
}

func getAcceptLanguage(r *http.Request) (langs []string) {
	for _, lang := range r.Header["Accept-Language"] {
		right := strings.Index(lang, ";")
		if right == -1 {
			right = len(lang)
		}
		l := strings.TrimSpace(lang[:right])
		if l == "" || l == "*" {
			continue
		}
		langs = append(langs, l)
	}
	return
}
