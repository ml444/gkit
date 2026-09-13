package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	jsoncodec "github.com/ml444/gkit/transport/httpx/coder/json"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func jsonTestRequest(rc IRouterCoder) *http.Request {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept", "application/xml")
	return r.WithContext(context.WithValue(r.Context(), routerCoderKey{}, rc))
}

func TestJSONCodecModeAndReset(t *testing.T) {
	opts := jsoncodec.DefaultOptions()
	opts.Marshal.UseProtoNames = true
	opts.Marshal.UseEnumNumbers = true
	rc, err := NewRouterCoder(WithJSONMode(JSONCodec), WithJSONCoder(jsoncodec.NewCoder(opts)))
	if err != nil {
		t.Fatal(err)
	}
	r := jsonTestRequest(rc)
	w := httptest.NewRecorder()
	c := NewCtx(w, r)
	value := &descriptorpb.FieldDescriptorProto{JsonName: proto.String("field"), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()}
	if err := c.JSON(201, value); err != nil {
		t.Fatal(err)
	}
	var fields map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &fields); err != nil {
		t.Fatal(err)
	}
	if w.Code != 201 || w.Header().Get("Content-Type") != "application/json" || fields["json_name"] != "field" || fields["type"] != float64(9) {
		t.Fatalf("JSON profile/Accept: %d %s", w.Code, w.Body.String())
	}
	response := httptest.NewRecorder()
	r.Header.Set("Accept", "application/json")
	if err := rc.ResponseEncoder()(201, response, r, value); err != nil {
		t.Fatal(err)
	}
	if w.Body.String() != response.Body.String() {
		t.Fatal("JSON and default JSON response disagree")
	}
	legacy := httptest.NewRecorder()
	c.Reset(legacy, httptest.NewRequest("GET", "/", nil))
	if err := c.JSON(200, map[string]int{"n": 1}); err != nil {
		t.Fatal(err)
	}
	if legacy.Body.String() != "{\"n\":1}\n" {
		t.Fatalf("Reset retained codec: %q", legacy.Body.String())
	}
	if _, err := NewRouterCoder(WithJSONMode(99)); err == nil {
		t.Fatal("invalid mode accepted")
	}
}

func TestJSONCodecEmptyAndWriteErrors(t *testing.T) {
	rc, _ := NewRouterCoder(WithJSONMode(JSONCodec))
	for _, tc := range []struct {
		method string
		status int
	}{{"HEAD", 200}, {"GET", 204}, {"GET", 205}, {"GET", 304}} {
		r := jsonTestRequest(rc)
		r.Method = tc.method
		w := httptest.NewRecorder()
		if err := NewCtx(w, r).JSON(tc.status, make(chan int)); err != nil || w.Body.Len() != 0 {
			t.Fatalf("bodyless JSON: %v %q", err, w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	if err := NewCtx(w, jsonTestRequest(rc)).JSON(200, nil); err != nil || w.Body.String() != "null" {
		t.Fatalf("JSON nil: %q %v", w.Body.String(), err)
	}
	w = httptest.NewRecorder()
	c := NewCtx(w, jsonTestRequest(rc))
	err := c.JSON(200, make(chan int))
	if err == nil || w.Body.Len() != 0 || w.Header().Get("Content-Type") != "" {
		t.Fatal("marshal failure committed a response")
	}
	c.ReturnError(err)
	if w.Code != 500 {
		t.Fatal(w.Code)
	}
	cause := errors.New("broken writer")
	fail := &failedResponseWriter{header: make(http.Header), err: cause}
	c = NewCtx(fail, jsonTestRequest(rc))
	err = c.JSON(200, struct{}{})
	if !errors.Is(err, cause) {
		t.Fatal(err)
	}
	c.ReturnError(err)
	if fail.headers != 1 || fail.writes != 1 {
		t.Fatalf("second response: %+v", fail)
	}
}

// External coders can still implement exactly the original six methods.
type sixMethodCoder struct{ IRouterCoder }

func TestJSONLegacyExternalCoderAndExplicitChoice(t *testing.T) {
	rc, _ := NewRouterCoder(WithJSONMode(JSONCodec), WithJSONCoder(markerCodec{"json", "instance"}),
		WithResponseEncoder(func(int, http.ResponseWriter, *http.Request, interface{}) error { return errors.New("must not run") }))
	w := httptest.NewRecorder()
	if err := NewCtx(w, jsonTestRequest(rc)).JSON(200, struct{}{}); err != nil || !strings.Contains(w.Body.String(), "instance") {
		t.Fatal(w.Body.String(), err)
	}
	w = httptest.NewRecorder()
	if err := NewCtx(w, jsonTestRequest(sixMethodCoder{rc})).JSON(200, struct{}{}); err != nil || w.Body.String() != "{}\n" {
		t.Fatal(w.Body.String(), err)
	}
}
