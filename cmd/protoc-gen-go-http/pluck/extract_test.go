package pluck

import (
	"net/http"
	"reflect"
	"testing"
)

type testReq struct {
	header     *testHeader
	mapHeader  map[string]string
	mapHeader2 map[string][]string
	data       []byte
}
type testHeader struct {
	ContentType     string `json:"Content-Type"`
	ContentLength   string `json:"Content-Length"`
	ContentEncoding string `json:"Content-Encoding"`
	XRequestId      string `json:"x-request-id"`
	XTracingId      string `json:"X-Tracing-Id"`
}

func TestExtractHeader(t *testing.T) {
	type args struct {
		header http.Header
		v      interface{}
	}
	var header *testHeader
	req := &testReq{}
	tests := []struct {
		name       string
		args       args
		checkValue interface{}
		wantErr    bool
	}{
		{
			name: "test-map-ok",
			args: args{
				header: map[string][]string{"Content-Type": {"application/json", "application/xml"}},
				v:      &req.mapHeader,
			},
			checkValue: &map[string]string{"Content-Type": "application/json,application/xml"},
			wantErr:    false,
		},
		{
			name: "test-struct-ok",
			args: args{
				header: map[string][]string{
					"Content-Type":     {"application/json", "application/xml"},
					"Content-Length":   {"100"},
					"Content-Encoding": {"gzip"},
					"x-request-id":     {"1234567890"},
					"X-Tracing-Id":     {"1234567890"},
				},
				v: &header,
			},
			checkValue: &testHeader{
				ContentType:     "application/json,application/xml",
				ContentLength:   "100",
				ContentEncoding: "gzip",
				XRequestId:      "1234567890",
				XTracingId:      "1234567890",
			},
			wantErr: false,
		},
		{
			name: "test-struct-nil",
			args: args{
				header: map[string][]string{
					"Content-Type":     {"application/json", "application/xml"},
					"Content-Length":   {"100"},
					"Content-Encoding": {"gzip"},
					"x-request-id":     {"1234567890"},
					"X-Tracing-Id":     {"1234567890"},
				},
				v: &req.header,
			},
			checkValue: &testHeader{
				ContentType:     "application/json,application/xml",
				ContentLength:   "100",
				ContentEncoding: "gzip",
				XRequestId:      "1234567890",
				XTracingId:      "1234567890",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ExtractHeader(tt.args.header, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("ExtractHeader() error = %v, wantErr %v", err, tt.wantErr)
			}
			got := tt.args.v
			if g, ok := got.(**testHeader); ok {
				got = *g
			}
			t.Logf("===> header: %v \n", got)
			if !reflect.DeepEqual(got, tt.checkValue) {
				t.Errorf("ExtractHeader() got = %v, want %v", tt.args.v, tt.checkValue)
			}
		})
	}
}

type testResponseWriter struct {
	header http.Header
}

func newTestResponseWriter() *testResponseWriter {
	return &testResponseWriter{
		header: http.Header{},
	}
}

func (t *testResponseWriter) Header() http.Header       { return t.header }
func (t *testResponseWriter) Write([]byte) (int, error) { return 0, nil }
func (t *testResponseWriter) WriteHeader(int)           {}

func TestSetResponseHeaders(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		h interface{}
	}
	req := testReq{
		header: &testHeader{
			ContentType: "application/json,application/xml",
		},
		mapHeader:  map[string]string{"Content-Type": "application/json,application/xml"},
		mapHeader2: map[string][]string{"Content-Type": {"application/json", "application/xml"}},
		data:       nil,
	}
	checkRspHeader := map[string][]string{"Content-Type": {"application/json,application/xml"}}
	tests := []struct {
		name    string
		args    args
		check   map[string][]string
		wantErr bool
	}{
		{
			name:    "test-map1",
			args:    args{w: newTestResponseWriter(), h: req.mapHeader},
			check:   checkRspHeader,
			wantErr: false,
		},
		{
			name:    "test-map2",
			args:    args{w: newTestResponseWriter(), h: req.mapHeader2},
			check:   checkRspHeader,
			wantErr: false,
		},
		{
			name:    "test-struct",
			args:    args{w: newTestResponseWriter(), h: req.header},
			check:   checkRspHeader,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetResponseHeaders(tt.args.w, tt.args.h); (err != nil) != tt.wantErr {
				t.Errorf("SetResponseHeaders() error: %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.w.Header(), http.Header(tt.check)) {
				t.Errorf("SetResponseHeaders() got: %v, want: %v", tt.args.w.Header(), tt.check)
			}
		})
	}
}

func TestConvertAnyToHeader(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantHeader http.Header
		wantErr    bool
	}{
		{
			name:       "test-map1",
			args:       args{v: map[string]string{"Content-Type": "application/json,application/xml"}},
			wantHeader: http.Header{"Content-Type": {"application/json,application/xml"}},
			wantErr:    false,
		},
		{
			name:       "test-map1-ptr",
			args:       args{v: &map[string]string{"Content-Type": "application/json,application/xml"}},
			wantHeader: http.Header{"Content-Type": {"application/json,application/xml"}},
			wantErr:    false,
		},
		{
			name:       "test-map2",
			args:       args{v: map[string][]string{"Content-Type": {"application/json", "application/xml"}}},
			wantHeader: http.Header{"Content-Type": {"application/json", "application/xml"}},
			wantErr:    false,
		},
		{
			name:       "test-struct",
			args:       args{v: &testHeader{ContentType: "application/json,application/xml"}},
			wantHeader: http.Header{"Content-Type": {"application/json,application/xml"}},
			wantErr:    false,
		},
		{
			name: "test-struct-[]string",
			args: args{v: &struct {
				ContentType []string `json:"Content-Type"`
			}{ContentType: []string{"application/json", "application/xml"}}},
			wantHeader: http.Header{"Content-Type": {"application/json", "application/xml"}},
			wantErr:    false,
		},
		{
			name:       "test-struct-nil",
			args:       args{v: &testHeader{}},
			wantHeader: http.Header{},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHeader, err := ConvertAnyToHeader(tt.args.v, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertAnyToHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotHeader, tt.wantHeader) {
				t.Errorf("ConvertAnyToHeader() gotHeader = %v, want %v", gotHeader, tt.wantHeader)
			}
		})
	}
}
