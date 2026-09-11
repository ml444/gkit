package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/mux"
	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/transport/httpx/coder"
	jsoncodec "github.com/ml444/gkit/transport/httpx/coder/json"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestRouterCoderPartialOverrideKeepsDefaults(t *testing.T) {
	want := errors.New("custom body decoder")
	rc, err := NewRouterCoder(WithBindBody(func(*http.Request, interface{}) error { return want }))
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, "/?value=query", strings.NewReader("value=form"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r = mux.SetURLVars(r, map[string]string{"value": "path"})
	for _, tc := range []struct {
		fn   RequestDecoder
		want string
	}{{rc.BindVars(), "path"}, {rc.BindQuery(), "query"}, {rc.BindForm(), "form"}} {
		var out struct {
			Value string `json:"value"`
		}
		if err := tc.fn(r, &out); err != nil || out.Value != tc.want {
			t.Fatalf("default binding: %+v, %v; want %q", out, err, tc.want)
		}
	}
	if err := rc.BindBody()(r, nil); !errors.Is(err, want) {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	if err := rc.ResponseEncoder()(201, w, r, map[string]bool{"ok": true}); err != nil || w.Code != 201 || w.Body.String() != `{"ok":true}` {
		t.Fatalf("default response: %d %s %v", w.Code, w.Body.String(), err)
	}
	w = httptest.NewRecorder()
	rc.ErrorEncoder()(w, r, errorx.BadRequest("bad"))
	if w.Code != 400 || !strings.Contains(w.Body.String(), "bad") {
		t.Fatalf("default error: %d %s", w.Code, w.Body.String())
	}
}

func TestRouterCoderAllOptionsAndOrder(t *testing.T) {
	first, last := errors.New("first"), errors.New("last")
	decoder := func(err error) RequestDecoder {
		return func(*http.Request, interface{}) error { return err }
	}
	for _, option := range []func(RequestDecoder) RouterCoderOption{WithBindVars, WithBindQuery, WithBindForm, WithBindBody} {
		rc, err := NewRouterCoder(option(decoder(first)), option(decoder(last)))
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, fn := range []RequestDecoder{rc.BindVars(), rc.BindQuery(), rc.BindForm(), rc.BindBody()} {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			var out struct{}
			if errors.Is(fn(r, &out), last) {
				found = true
			}
		}
		if !found {
			t.Fatal("later decoder did not override earlier option")
		}
	}
	response := func(err error) ResponseEncoder {
		return func(int, http.ResponseWriter, *http.Request, interface{}) error { return err }
	}
	errorEncoder := func(code int) ErrorEncoder {
		return func(w http.ResponseWriter, _ *http.Request, _ error) { w.WriteHeader(code) }
	}
	base, err := NewRouterCoder(WithBindBody(decoder(first)), WithResponseEncoder(response(first)), WithErrorEncoder(errorEncoder(401)))
	if err != nil {
		t.Fatal(err)
	}
	rc, err := NewRouterCoder(WithBaseRouterCoder(base), WithBindBody(decoder(last)),
		WithResponseEncoder(response(last)), WithErrorEncoder(errorEncoder(403)))
	if err != nil {
		t.Fatal(err)
	}
	if err := rc.BindBody()(nil, nil); !errors.Is(err, last) {
		t.Fatal(err)
	}
	if err := rc.ResponseEncoder()(200, nil, nil, nil); !errors.Is(err, last) {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	rc.ErrorEncoder()(w, nil, last)
	if w.Code != 403 {
		t.Fatal(w.Code)
	}
	rc, err = NewRouterCoder(WithBindBody(decoder(last)), WithBaseRouterCoder(base))
	if err != nil || !errors.Is(rc.BindBody()(nil, nil), first) {
		t.Fatalf("base applied last must replace callbacks: %v", err)
	}
}

func TestRouterCoderRejectsInvalidOptions(t *testing.T) {
	invalidBase := newRouterCoder()
	invalidBase.bindQuery = nil
	for _, option := range []RouterCoderOption{
		nil, WithBindVars(nil), WithBindQuery(nil), WithBindForm(nil), WithBindBody(nil),
		WithResponseEncoder(nil), WithErrorEncoder(nil), WithBaseRouterCoder(nil),
		WithBaseRouterCoder((*routerCoder)(nil)), WithBaseRouterCoder(invalidBase),
		WithJSONCoder(nil), WithJSONCoder((*jsoncodec.ConfiguredCoder)(nil)),
		WithJSONCoder(markerCodec{name: "xml"}),
	} {
		if rc, err := NewRouterCoder(option); err == nil || rc != nil {
			t.Fatalf("accepted invalid option: %v, %v", rc, err)
		}
	}
}

type markerCodec struct {
	name, marker string
}

func (c markerCodec) Name() string { return c.name }
func (c markerCodec) Marshal(interface{}) ([]byte, error) {
	return json.Marshal(map[string]string{"marker": c.marker})
}
func (c markerCodec) Unmarshal(_ []byte, v interface{}) error {
	*v.(*string) = c.marker
	return nil
}

func TestRouterCoderJSONReplacementAndCustomCallbacks(t *testing.T) {
	rc, err := NewRouterCoder(WithJSONCoder(markerCodec{"json", "first"}), WithJSONCoder(markerCodec{"json", "last"}))
	if err != nil {
		t.Fatal(err)
	}
	for _, media := range []string{"application/json", "application/not-supported", ""} {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		r.Header.Set("Content-Type", media)
		r.Header.Set("Accept", media)
		var out string
		if err := rc.BindBody()(r, &out); err != nil || out != "last" {
			t.Fatalf("JSON binding/fallback: %q %v", out, err)
		}
		w := httptest.NewRecorder()
		if err := rc.ResponseEncoder()(200, w, r, struct{}{}); err != nil || !strings.Contains(w.Body.String(), "last") {
			t.Fatalf("JSON response/fallback: %s %v", w.Body.String(), err)
		}
		w = httptest.NewRecorder()
		rc.ErrorEncoder()(w, r, errorx.BadRequest("bad"))
		if w.Code != 400 || !strings.Contains(w.Body.String(), "last") {
			t.Fatalf("JSON error/fallback: %d %s", w.Code, w.Body.String())
		}
	}
	base, err := NewRouterCoder(WithJSONCoder(markerCodec{"json", "base"}))
	if err != nil {
		t.Fatal(err)
	}
	derived, err := NewRouterCoder(WithBaseRouterCoder(base), WithJSONCoder(markerCodec{"json", "derived"}))
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	if err := derived.ResponseEncoder()(200, w, httptest.NewRequest("GET", "/", nil), struct{}{}); err != nil || !strings.Contains(w.Body.String(), "base") {
		t.Fatalf("base callback's codec must remain unchanged: %s %v", w.Body.String(), err)
	}
}

func TestRouterCoderRegistrySnapshotAndLegacyLookup(t *testing.T) {
	const name = "profile-snapshot-test"
	if err := coder.RegisterCoder(markerCodec{name, "before"}); err != nil {
		t.Fatal(err)
	}
	rc, err := NewRouterCoder()
	if err != nil {
		t.Fatal(err)
	}
	if err := coder.RegisterCoder(markerCodec{name, "after"}); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept", "application/"+name)
	for _, tc := range []struct {
		c    IRouterCoder
		want string
	}{{rc, "before"}, {newRouterCoder(), "after"}} {
		w := httptest.NewRecorder()
		if err := tc.c.ResponseEncoder()(200, w, r, struct{}{}); err != nil || !strings.Contains(w.Body.String(), tc.want) {
			t.Fatalf("registry behavior: %s %v, want %q", w.Body.String(), err, tc.want)
		}
	}
}

func TestRouterCoderSnapshotsBuiltinJSONButPreservesCustomJSON(t *testing.T) {
	original, _ := coder.LookupCoder("json")
	oldOptions := jsoncodec.MarshalOptions
	t.Cleanup(func() {
		jsoncodec.MarshalOptions = oldOptions
		if err := coder.RegisterCoder(original); err != nil {
			t.Error(err)
		}
	})
	jsoncodec.MarshalOptions.UseProtoNames = true
	rc, err := NewRouterCoder()
	if err != nil {
		t.Fatal(err)
	}
	jsoncodec.MarshalOptions.UseProtoNames = false
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	if err := rc.ResponseEncoder()(200, w, r, &descriptorpb.FieldDescriptorProto{JsonName: proto.String("name")}); err != nil || !strings.Contains(w.Body.String(), `"json_name"`) {
		t.Fatalf("builtin JSON was not frozen: %s %v", w.Body.String(), err)
	}
	if err := coder.RegisterCoder(markerCodec{"json", "registered-custom"}); err != nil {
		t.Fatal(err)
	}
	rc, err = NewRouterCoder()
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	if err := rc.ResponseEncoder()(200, w, r, struct{}{}); err != nil || !strings.Contains(w.Body.String(), "registered-custom") {
		t.Fatalf("registered custom JSON was replaced: %s %v", w.Body.String(), err)
	}
}

func TestRouterCoderServerInstancesConcurrent(t *testing.T) {
	servers := make([]*Server, 2)
	for i := range servers {
		opts := jsoncodec.DefaultOptions()
		opts.Marshal.UseProtoNames = i == 0
		rc, err := NewRouterCoder(WithJSONCoder(jsoncodec.NewCoder(opts)))
		if err != nil {
			t.Fatal(err)
		}
		servers[i] = NewServer(RouterCoder(rc))
		servers[i].GetRouter().GET("/", func(w http.ResponseWriter, r *http.Request) {
			NewCtx(w, r).Result(200, &descriptorpb.FieldDescriptorProto{JsonName: proto.String("name")})
		})
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				for k, srv := range servers {
					w := httptest.NewRecorder()
					srv.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
					key := `"jsonName"`
					if k == 0 {
						key = `"json_name"`
					}
					if w.Code != 200 || !strings.Contains(w.Body.String(), key) {
						t.Errorf("instance %d: %d %s", k, w.Code, w.Body.String())
					}
				}
			}
		}()
	}
	wg.Wait()
}
