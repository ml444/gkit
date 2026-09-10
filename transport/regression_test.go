package transport_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ml444/gkit/discovery"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/transport"
	"github.com/ml444/gkit/transport/grpcx"
	"github.com/ml444/gkit/transport/httpx"
	"github.com/ml444/gkit/transport/httpx/coder/form"
	"google.golang.org/grpc"
	hp "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/typepb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func response() *http.Response {
	return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{}`))}
}
func TestHTTPFullEndpoint(t *testing.T) {
	c, e := httpx.NewClient(httpx.WithEndpoint("http://example.test/api"), httpx.WithTransport(roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "example.test" {
			t.Errorf("full URL copied into host: %q", r.URL.Host)
		}
		if r.URL.Path != "/api/items" {
			t.Errorf("base path missing: %q", r.URL.Path)
		}
		return response(), nil
	})))
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	if e = c.Invoke(context.Background(), "GET", "/items", nil, nil); e != nil {
		t.Errorf("full URL invoke: %v", e)
	}
}
func TestHTTPDoHTTPS(t *testing.T) {
	c, e := httpx.NewClient(httpx.WithTransport(roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.URL.Scheme != "https" {
			t.Errorf("HTTPS request scheme changed to %q", r.URL.Scheme)
		}
		return response(), nil
	})))
	if e != nil {
		t.Fatal(e)
	}
	r, _ := http.NewRequest("GET", "https://example.test/", nil)
	res, e := c.Do(r)
	if e != nil {
		t.Fatal(e)
	}
	res.Body.Close()
}
func TestHTTPMiddlewareHeaders(t *testing.T) {
	mw := func(next middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, in interface{}) (interface{}, error) {
			tr, _ := transport.FromContext(ctx)
			tr.In().Set("Authorization", "Bearer test")
			return next(ctx, in)
		}
	}
	c, e := httpx.NewClient(httpx.WithEndpoint("example.test"), httpx.WithMiddlewares(mw), httpx.WithTransport(roundTrip(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Authorization") == "" {
			t.Error("middleware header missing from outbound request")
		}
		return response(), nil
	})))
	if e != nil {
		t.Fatal(e)
	}
	if e = c.Invoke(context.Background(), "GET", "/", nil, nil); e != nil {
		t.Fatal(e)
	}
}
func TestHTTP204(t *testing.T) {
	var out struct{}
	r := response()
	r.StatusCode = 204
	r.Body = io.NopCloser(strings.NewReader(""))
	if e := httpx.DefaultResponseDecoder(context.Background(), r, &out); e != nil {
		t.Errorf("204 considered error: %v", e)
	}
}
func TestHTTPGroupHandle(t *testing.T) {
	s := httpx.NewServer()
	auth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401) })
	}
	g := s.NewRouteGroup("/admin", auth)
	g.HandleFunc("/secret", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("secret")) })
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest("GET", "/secret", nil))
	if w.Code == 200 {
		t.Errorf("group route exposed outside prefix and auth: %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest("GET", "/admin/secret", nil))
	if w.Code != 401 {
		t.Errorf("group-prefixed route status = %d, want 401", w.Code)
	}
}
func TestHTTPGroupUseScope(t *testing.T) {
	s := httpx.NewServer()
	s.GetRouter().GET("/public", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	s.NewRouteGroup("/admin").Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(401) })
	})
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest("GET", "/public", nil))
	if w.Code != 200 {
		t.Errorf("group Use affects public route: %d", w.Code)
	}
}
func TestHTTPMultiValueHeaders(t *testing.T) {
	s := httpx.NewServer()
	s.GetRouter().GET("/", func(w http.ResponseWriter, r *http.Request) {
		tr, _ := transport.FromContext(r.Context())
		tr.Out().Append("Set-Cookie", "a=1", "b=2")
		w.WriteHeader(200)
	})
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if len(w.Result().Cookies()) != 2 {
		t.Errorf("cookies lost: %v", w.Header().Values("Set-Cookie"))
	}
}
func TestEncodeURLReserved(t *testing.T) {
	s := httpx.EncodeURL("/items/{id}", struct {
		ID string `json:"id"`
	}{"a?admin=true#frag"}, false)
	u, e := url.Parse(s)
	if e != nil {
		t.Fatal(e)
	}
	if u.RawQuery != "" || u.Fragment != "" {
		t.Errorf("path value changes URL structure: %s", s)
	}
}
func TestProtoBytesRoundTrip(t *testing.T) {
	m := wrapperspb.Bytes([]byte{0xfb, 0xff})
	q, e := form.EncodeValues(m)
	if e != nil {
		t.Fatal(e)
	}
	out := new(wrapperspb.BytesValue)
	if e = form.DecodeValues(out, q); e != nil {
		t.Errorf("cannot decode own encoded bytes %v: %v", q, e)
	}
}
func TestProtoNullTimestamp(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.Errorf("valid null timestamp causes panic: %v", p)
		}
	}()
	// A dynamic descriptor provides a Timestamp field without adding test-generated code.
	fd, e := protodesc.NewFile(&descriptorpb.FileDescriptorProto{Name: proto.String("null.proto"), Package: proto.String("review"), Syntax: proto.String("proto3"), Dependency: []string{"google/protobuf/timestamp.proto"}, MessageType: []*descriptorpb.DescriptorProto{{Name: proto.String("Request"), Field: []*descriptorpb.FieldDescriptorProto{{Name: proto.String("at"), Number: proto.Int32(1), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".google.protobuf.Timestamp")}}}}}, protoregistry.GlobalFiles)
	if e != nil {
		t.Fatal(e)
	}
	if e = form.DecodeValues(dynamicpb.NewMessage(fd.Messages().Get(0)), url.Values{"at": {"null"}}); e != nil {
		t.Errorf("null timestamp: %v", e)
	}
}
func TestUnknownEnumEncode(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.Errorf("unknown enum causes panic: %v", p)
		}
	}()
	_, e := form.EncodeValues(&typepb.Type{Syntax: typepb.Syntax(999)})
	if e != nil {
		t.Log(e)
	}
}
func TestProtoMap(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.Errorf("map encode panics: %v", p)
		}
	}()
	fd, e := protodesc.NewFile(&descriptorpb.FileDescriptorProto{Name: proto.String("map.proto"), Package: proto.String("review"), Syntax: proto.String("proto3"), MessageType: []*descriptorpb.DescriptorProto{{Name: proto.String("Request"), Field: []*descriptorpb.FieldDescriptorProto{{Name: proto.String("flags"), Number: proto.Int32(1), Label: descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".review.Request.FlagsEntry")}}, NestedType: []*descriptorpb.DescriptorProto{{Name: proto.String("FlagsEntry"), Options: &descriptorpb.MessageOptions{MapEntry: proto.Bool(true)}, Field: []*descriptorpb.FieldDescriptorProto{{Name: proto.String("key"), Number: proto.Int32(1), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_STRING.Enum()}, {Name: proto.String("value"), Number: proto.Int32(2), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_BOOL.Enum()}}}}}}}, nil)
	if e != nil {
		t.Fatal(e)
	}
	m := dynamicpb.NewMessage(fd.Messages().Get(0))
	m.Mutable(m.Descriptor().Fields().Get(0)).Map().Set(protoreflect.ValueOfString("enabled").MapKey(), protoreflect.ValueOfBool(true))
	if _, e = form.EncodeValues(m); e != nil {
		t.Errorf("map encode: %v", e)
	}
}

type slowHealth struct{ hp.UnimplementedHealthServer }

func (slowHealth) Check(ctx context.Context, r *hp.HealthCheckRequest) (*hp.HealthCheckResponse, error) {
	select {
	case <-time.After(60 * time.Millisecond):
		return &hp.HealthCheckResponse{}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
func TestGRPCDefaultTimeout(t *testing.T) {
	l := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	hp.RegisterHealthServer(s, slowHealth{})
	go s.Serve(l)
	defer s.Stop()
	c, e := grpcx.NewClient(grpcx.WithEndpoint("buf"), grpcx.WithTimeout(10*time.Millisecond), grpcx.WithDialOptions(grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return l.DialContext(ctx) })))
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	start := time.Now()
	_, e = hp.NewHealthClient(c.Conn()).Check(ctx, &hp.HealthCheckRequest{})
	if e == nil {
		t.Errorf("10ms default call timeout did not apply; success after %s", time.Since(start))
	}
}
func TestGRPCStopBeforeStart(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.Errorf("public NewServer -> Stop panics: %v", p)
		}
	}()
	s, e := grpcx.NewServer()
	if e != nil {
		t.Fatal(e)
	}
	s.Stop(context.Background())
}

func TestGRPCFeedback(t *testing.T) {
	l := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	hp.RegisterHealthServer(s, slowHealth{})
	go s.Serve(l)
	defer s.Stop()
	c, e := grpcx.NewClient(grpcx.WithEndpoint("buf"), grpcx.WithDiscovery(discovery.NewDiscoveryClient(discovery.NewDefaultRegistry()), "svc"), grpcx.WithDialOptions(grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return l.DialContext(ctx) })))
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, e = hp.NewHealthClient(c.Conn()).Check(ctx, &hp.HealthCheckRequest{}); e != nil {
		t.Errorf("successful RPC replaced by feedback error: %v", e)
	}
}

func TestFieldMaskMutation(t *testing.T) {
	fd, e := protodesc.NewFile(&descriptorpb.FileDescriptorProto{Name: proto.String("mask.proto"), Package: proto.String("review"), Syntax: proto.String("proto3"), Dependency: []string{"google/protobuf/field_mask.proto"}, MessageType: []*descriptorpb.DescriptorProto{{Name: proto.String("Request"), Field: []*descriptorpb.FieldDescriptorProto{{Name: proto.String("update_mask"), Number: proto.Int32(1), Label: descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(), Type: descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(), TypeName: proto.String(".google.protobuf.FieldMask")}}}}}, protoregistry.GlobalFiles)
	if e != nil {
		t.Fatal(e)
	}
	mask := &fieldmaskpb.FieldMask{Paths: []string{"user_name"}}
	m := dynamicpb.NewMessage(fd.Messages().Get(0))
	m.Set(m.Descriptor().Fields().Get(0), protoreflect.ValueOfMessage(mask.ProtoReflect()))
	if _, e = form.EncodeValues(m); e != nil {
		t.Fatal(e)
	}
	if mask.Paths[0] != "user_name" {
		t.Errorf("encoding mutated input FieldMask: %v", mask.Paths)
	}
}

type failingWriter struct{ header http.Header }

func (w *failingWriter) Header() http.Header     { return w.header }
func (*failingWriter) WriteHeader(int)           {}
func (*failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func TestResponseWriteError(t *testing.T) {
	w := &failingWriter{header: make(http.Header)}
	e := httpx.NewRouterCfg().Coder.ResponseEncoder()(200, w, httptest.NewRequest("GET", "/", nil), struct{}{})
	if e == nil {
		t.Error("response encoder hid writer failure")
	}
}
