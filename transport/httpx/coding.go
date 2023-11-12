package httpx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"

	"github.com/gorilla/mux"
	"google.golang.org/protobuf/proto"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/transport/httpx/encoding"
	"github.com/ml444/gkit/transport/httpx/encoding/form"
	"github.com/ml444/gkit/transport/httpx/encoding/json"
	encodingproto "github.com/ml444/gkit/transport/httpx/encoding/proto"
	"github.com/ml444/gkit/transport/httpx/encoding/xml"
)

var reg = regexp.MustCompile(`{[\\.\w]+}`)

// EncodeURL encode proto message to url path.
func EncodeURL(pathTemplate string, msg interface{}, needQuery bool) string {
	if msg == nil || (reflect.ValueOf(msg).Kind() == reflect.Ptr && reflect.ValueOf(msg).IsNil()) {
		return pathTemplate
	}
	queryParams, _ := form.EncodeValues(msg)
	pathParams := make(map[string]struct{})
	path := reg.ReplaceAllStringFunc(pathTemplate, func(in string) string {
		// it's unreachable because the reg means that must have more than one char in {}
		// if len(in) < 4 { //nolint:gomnd // **  explain the 4 number here :-) **
		//	return in
		// }
		key := in[1 : len(in)-1]
		pathParams[key] = struct{}{}
		return queryParams.Get(key)
	})
	if !needQuery {
		if v, ok := msg.(proto.Message); ok {
			if query := form.EncodeFieldMask(v.ProtoReflect()); query != "" {
				return path + "?" + query
			}
		}
		return path
	}
	if len(queryParams) > 0 {
		for key := range pathParams {
			delete(queryParams, key)
		}
		if query := queryParams.Encode(); query != "" {
			path += "?" + query
		}
	}
	return path
}

var codecInitFuncMap = map[string]func(){
	form.Name:          form.Init,
	json.Name:          json.Init,
	encodingproto.Name: encodingproto.Init,
	xml.Name:           xml.Init,
}

func getCoder(contentSubtype string) encoding.Coder {
	c := encoding.GetCoder(contentSubtype)
	if c == nil {
		if initFunc, ok := codecInitFuncMap[contentSubtype]; ok {
			initFunc()
			return encoding.GetCoder(contentSubtype)
		}
	}
	return c
}

// DecodeRequestFunc is decode request func.
type DecodeRequestFunc func(*http.Request, interface{}) error

// EncodeResponseFunc is encode response func.
type EncodeResponseFunc func(http.ResponseWriter, *http.Request, interface{}) error

// EncodeErrorFunc is encode error func.
type EncodeErrorFunc func(http.ResponseWriter, *http.Request, error)

// DefaultResponseEncoder encodes the object to the HTTP response.
func DefaultResponseEncoder(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if v == nil {
		return nil
	}
	if rd, ok := v.(IRedirect); ok {
		redirectUrl, code := rd.Redirect()
		http.Redirect(w, r, redirectUrl, code)
		return nil
	}
	codec, _ := CodecForRequest(r, "Accept")
	data, err := codec.Marshal(v)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", JoinContentType(codec.Name()))
	_, err = w.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// DefaultErrorEncoder encodes the error to the HTTP response.
func DefaultErrorEncoder(w http.ResponseWriter, r *http.Request, err error) {
	se := errorx.FromError(err)
	codec, _ := CodecForRequest(r, "Accept")
	body, err := codec.Marshal(se)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", JoinContentType(codec.Name()))
	w.WriteHeader(int(se.StatusCode))
	_, _ = w.Write(body)
}

// DefaultRequestVars decodes the request vars to object.
func DefaultRequestVars(r *http.Request, target interface{}) error {
	raws := mux.Vars(r)
	vars := make(url.Values, len(raws))
	for k, v := range raws {
		vars[k] = []string{v}
	}
	if err := getCoder(form.Name).Unmarshal([]byte(vars.Encode()), target); err != nil {
		return defaultError(err)
	}
	return nil
}

// DefaultRequestDecoder decodes the request body to object.
func DefaultRequestDecoder(r *http.Request, v interface{}) error {
	codec, ok := CodecForRequest(r, "Content-Type")
	if !ok {
		return errorx.BadRequest(fmt.Sprintf("unregister Content-Type: %s", r.Header.Get("Content-Type")))
	}
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

// CodecForRequest get encoding.Coder via http.Request
func CodecForRequest(r *http.Request, name string) (encoding.Coder, bool) {
	for _, accept := range r.Header[name] {
		codec := getCoder(ContentSubtype(accept))
		if codec != nil {
			return codec, true
		}
	}
	return getCoder("json"), false
}

// DefaultRequestEncoder is an HTTP request encoder.
func DefaultRequestEncoder(_ context.Context, contentType string, in interface{}) ([]byte, error) {
	name := ContentSubtype(contentType)
	body, err := getCoder(name).Marshal(in)
	if err != nil {
		return nil, err
	}
	return body, err
}

// DefaultResponseDecoder is an HTTP response decoder.
func DefaultResponseDecoder(_ context.Context, rsp *http.Response, v interface{}) error {
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
		}
	}
	return coderByContentType(rsp.Header.Get("Content-Type")).Unmarshal(data, v)
}

func coderByContentType(contentType string) encoding.Coder {
	codec := getCoder(ContentSubtype(contentType))
	if codec != nil {
		return codec
	}
	return getCoder("json")
}
