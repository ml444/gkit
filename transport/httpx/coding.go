package httpx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/transport/httpx/coder"
	"github.com/ml444/gkit/transport/httpx/coder/form"
)

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

	//SetBindVars(RequestDecoder)
	//SetBindQuery(RequestDecoder)
	//SetBindForm(RequestDecoder)
	//SetBindBody(RequestDecoder)
	//SetResponseEncoder(ResponseEncoder)
	//SetErrorEncoder(ErrorEncoder)
}

type routerCoder struct{}

func (c *routerCoder) BindVars() RequestDecoder {
	return func(r *http.Request, target interface{}) error {
		raws := mux.Vars(r)
		vars := make(url.Values, len(raws))
		for k, v := range raws {
			vars[k] = []string{v}
		}
		if err := coder.GetCoder(form.Name).Unmarshal([]byte(vars.Encode()), target); err != nil {
			return defaultError(err)
		}
		return nil
	}
}
func (c *routerCoder) BindQuery() RequestDecoder {
	return func(r *http.Request, v interface{}) error {
		if err := coder.GetCoder(form.Name).Unmarshal([]byte(r.URL.Query().Encode()), v); err != nil {
			return defaultError(err)
		}
		return nil
	}

}
func (c *routerCoder) BindForm() RequestDecoder {
	return func(r *http.Request, v interface{}) error {
		if err := r.ParseForm(); err != nil {
			return err
		}
		if err := coder.GetCoder(form.Name).Unmarshal([]byte(r.Form.Encode()), v); err != nil {
			return defaultError(err)
		}
		return nil
	}

}
func (c *routerCoder) BindBody() RequestDecoder {
	return func(r *http.Request, v interface{}) error {
		codec, _ := coderForRequest(r, "Content-Type")
		//if !ok {
		//	return errorx.BadRequest(fmt.Sprintf("unregister Content-Type: %s", r.Header.Get("Content-Type")))
		//}
		data, err := io.ReadAll(r.Body)

		// reset body.
		r.Body = io.NopCloser(bytes.NewBuffer(data))

		if err != nil {
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

}

func (c *routerCoder) ResponseEncoder() ResponseEncoder {
	return func(status int, w http.ResponseWriter, r *http.Request, v interface{}) error {
		if v == nil {
			return nil
		}
		if rd, ok := v.(IRedirect); ok {
			redirectUrl, code := rd.Redirect()
			http.Redirect(w, r, redirectUrl, code)
			return nil
		}
		codec, _ := coderForRequest(r, "Accept")
		data, err := codec.Marshal(v)
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", joinContentType(codec.Name()))
		w.WriteHeader(status)
		_, err = w.Write(data)
		if err != nil {
			return err
		}
		return nil
	}
}

func (c *routerCoder) ErrorEncoder() ErrorEncoder {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		se := errorx.FromError(err)
		codec, _ := coderForRequest(r, "Accept")
		body, err := codec.Marshal(se)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", joinContentType(codec.Name()))
		w.WriteHeader(int(se.StatusCode))
		_, _ = w.Write(body)
	}
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
	if rsp.StatusCode < 400 && v == nil {
		return nil
	}
	defer rsp.Body.Close()
	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	if rsp.StatusCode >= 400 {
		e := new(errorx.Error)
		if err = coderByContentType(rsp.Header.Get("Content-Type")).Unmarshal(data, e); err == nil {
			e.StatusCode = int32(rsp.StatusCode)
			return e
		} else {
			e.StatusCode = int32(rsp.StatusCode)
			e.ErrorCode = errorx.ErrCodeInvalidReqSys
			e.Message = string(data)
			return e
		}
	}
	return coderByContentType(rsp.Header.Get("Content-Type")).Unmarshal(data, v)
}

func coderForRequest(r *http.Request, name string) (coder.ICoder, bool) {
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
