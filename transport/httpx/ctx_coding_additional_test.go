package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/transport"
)

func TestWrappedCtxAccessorsAndBinders(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/users/42?q=neo", strings.NewReader(`{"name":"body"}`))
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "42"})
	rec := httptest.NewRecorder()
	ctx := NewCtx(rec, req)

	if ctx.Header().Get("Content-Type") == "" || ctx.Request() != req || ctx.Response() != rec {
		t.Fatalf("unexpected accessors")
	}
	if ctx.Vars().Get("id") != "42" || ctx.Query().Get("q") != "neo" {
		t.Fatalf("vars/query mismatch")
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := ctx.Bind(&body); err != nil || body.Name != "body" {
		t.Fatalf("bind body = %#v %v", body, err)
	}
	var vars struct {
		ID string `json:"id"`
	}
	if err := ctx.BindVars(&vars); err != nil || vars.ID != "42" {
		t.Fatalf("bind vars = %#v %v", vars, err)
	}
	var query struct {
		Q string `json:"q"`
	}
	if err := ctx.BindQuery(&query); err != nil || query.Q != "neo" {
		t.Fatalf("bind query = %#v %v", query, err)
	}

	formReq := httptest.NewRequest(http.MethodPost, "/form", strings.NewReader("name=form"))
	formReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	ctx.Reset(rec, formReq)
	if ctx.Form().Get("name") != "form" {
		t.Fatalf("form = %#v", ctx.Form())
	}
	var form struct {
		Name string `json:"name"`
	}
	if err := ctx.BindForm(&form); err != nil || form.Name != "form" {
		t.Fatalf("bind form = %#v %v", form, err)
	}
}

func TestWrappedCtxResponseMethods(t *testing.T) {
	methods := []struct {
		name string
		run  func(Context) error
	}{
		{"JSON", func(c Context) error { return c.JSON(http.StatusCreated, map[string]string{"ok": "1"}) }},
		{"XML", func(c Context) error { return c.XML(http.StatusCreated, xmlPayload{V: "x"}) }},
		{"String", func(c Context) error { return c.String(http.StatusCreated, "text") }},
		{"Blob", func(c Context) error { return c.Blob(http.StatusCreated, "text/plain", []byte("blob")) }},
		{"Stream", func(c Context) error { return c.Stream(http.StatusCreated, "text/plain", strings.NewReader("stream")) }},
	}
	for _, tt := range methods {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			tr := &Transport{outMD: transport.Pairs("X-From-Transport", "1")}
			req = req.WithContext(transport.ToContext(req.Context(), tr))
			rec := httptest.NewRecorder()
			if err := tt.run(NewCtx(rec, req)); err != nil {
				t.Fatalf("run: %v", err)
			}
			if rec.Code != http.StatusCreated || rec.Header().Get("X-From-Transport") != "1" {
				t.Fatalf("response code/header = %d %#v", rec.Code, rec.Header())
			}
		})
	}
}

func TestWrappedCtxResultReturnsAndError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	ctx := NewCtx(rec, req)
	ctx.Returns(map[string]string{"ok": "1"}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("returns code = %d", rec.Code)
	}
	var decoded map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil || decoded["ok"] != "1" {
		t.Fatalf("body = %q %v", rec.Body.String(), err)
	}

	rec = httptest.NewRecorder()
	ctx.Reset(rec, req)
	ctx.Returns(nil, errorx.BadRequest("bad"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("error code = %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	badCoder := newRouterCoder()
	badCoder.respEnc = func(int, http.ResponseWriter, *http.Request, interface{}) error {
		return errors.New("encode")
	}
	reqWithCoder := req.WithContext(context.WithValue(req.Context(), routerCoderKey{}, badCoder))
	NewCtx(rec, reqWithCoder).Result(http.StatusOK, map[string]string{"bad": "1"})
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("result error code = %d", rec.Code)
	}
}

func TestWrappedCtxContextMethods(t *testing.T) {
	base, cancel := context.WithTimeout(context.WithValue(context.Background(), "k", "v"), time.Minute)
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(base)
	ctx := NewCtx(httptest.NewRecorder(), req)
	if _, ok := ctx.Deadline(); !ok {
		t.Fatal("expected deadline")
	}
	if ctx.Done() == nil || ctx.Err() != nil || ctx.Value("k") != "v" {
		t.Fatalf("context methods failed")
	}

	empty := &wrappedCtx{}
	if _, ok := empty.Deadline(); ok || empty.Done() != nil || !errors.Is(empty.Err(), context.Canceled) || empty.Value("k") != nil {
		t.Fatalf("empty context methods failed")
	}
}

func TestRouterCoderHelpers(t *testing.T) {
	c := newRouterCoder()
	if c.BindVars() == nil || c.BindQuery() == nil || c.BindForm() == nil || c.BindBody() == nil ||
		c.ResponseEncoder() == nil || c.ErrorEncoder() == nil {
		t.Fatal("expected coder funcs")
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if codec, ok := getCoderForRequest(req, "Accept"); ok || codec == nil {
		t.Fatalf("default request coder = %#v %v", codec, ok)
	}
	req.Header.Set("Accept", "application/xml")
	if codec, ok := getCoderForRequest(req, "Accept"); !ok || codec.Name() != "xml" {
		t.Fatalf("xml request coder = %#v %v", codec, ok)
	}
	if got := contentSubtype("application/json; charset=utf-8"); got != "json" {
		t.Fatalf("subtype = %q", got)
	}
	if got := contentSubtype("plain"); got != "" {
		t.Fatalf("plain subtype = %q", got)
	}
	if got := coderByContentType("unknown/unknown").Name(); got != "json" {
		t.Fatalf("fallback content coder = %q", got)
	}
	if err := defaultError(url.EscapeError("bad")); err == nil {
		t.Fatal("expected default error")
	}
	data, err := DefaultRequestEncoder(context.Background(), "application/json", map[string]string{"a": "b"})
	if err != nil || len(data) == 0 {
		t.Fatalf("default request encode = %q %v", data, err)
	}
}

type xmlPayload struct {
	V string `xml:"v"`
}

func TestDefaultResponseDecoderBranches(t *testing.T) {
	if err := DefaultResponseDecoder(context.Background(), &http.Response{
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     http.Header{},
	}, nil); err != nil {
		t.Fatalf("nil success decode: %v", err)
	}
	var out map[string]string
	if err := DefaultResponseDecoder(context.Background(), &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"a":"b"}`)),
		Header:     http.Header{"Content-Type": {"application/json"}},
	}, &out); err != nil || out["a"] != "b" {
		t.Fatalf("success decode = %#v %v", out, err)
	}
	err := DefaultResponseDecoder(context.Background(), &http.Response{
		StatusCode: http.StatusTeapot,
		Body:       io.NopCloser(strings.NewReader("plain error")),
		Header:     http.Header{"Content-Type": {"text/plain"}},
	}, &out)
	if errorx.Status(err) != http.StatusTeapot {
		t.Fatalf("plain error decode = %v", err)
	}
}
