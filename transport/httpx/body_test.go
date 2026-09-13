package httpx

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/transport/httpx/coder"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestReadRequestBodyLengthHints(t *testing.T) {
	payload := strings.Repeat("x", 64<<10)
	for _, hint := range []int64{-1, 0, 1, 511, 512, int64(len(payload)), math.MaxInt64} {
		t.Run(fmt.Sprint(hint), func(t *testing.T) {
			got, err := readRequestBody(strings.NewReader(payload), hint)
			if err != nil || string(got) != payload {
				t.Fatalf("read %d bytes: %v", len(got), err)
			}
			if hint == math.MaxInt64 && cap(got) > maxRequestBodyPrealloc+bytes.MinRead {
				t.Fatalf("unbounded speculative allocation: capacity %d", cap(got))
			}
		})
	}
	// The allocation hint is not a limit on the actual body size.
	payload = strings.Repeat("x", 2*maxRequestBodyPrealloc)
	got, err := readRequestBody(strings.NewReader(payload), int64(len(payload)))
	if err != nil || string(got) != payload {
		t.Fatalf("body beyond hint cap: %d bytes, %v", len(got), err)
	}
}

func TestReadRequestBodyReaderBehavior(t *testing.T) {
	for _, hint := range []int64{-1, 1024} {
		for _, tc := range []struct {
			name string
			body io.Reader
			want string
			err  error
		}{
			{"empty", strings.NewReader(""), "", nil},
			{"short-reads", iotest.OneByteReader(strings.NewReader("body")), "body", nil},
			{"data-and-EOF", iotest.DataErrReader(strings.NewReader("body")), "body", nil},
			{"truncated", io.MultiReader(strings.NewReader("body"), iotest.ErrReader(io.ErrUnexpectedEOF)), "body", io.ErrUnexpectedEOF},
		} {
			t.Run(fmt.Sprintf("%d/%s", hint, tc.name), func(t *testing.T) {
				got, err := readRequestBody(tc.body, hint)
				if string(got) != tc.want || !errors.Is(err, tc.err) {
					t.Fatalf("read = %q, %v", got, err)
				}
			})
		}
		limited := http.MaxBytesReader(httptest.NewRecorder(), io.NopCloser(strings.NewReader("large")), 3)
		got, err := readRequestBody(limited, hint)
		var mbe *http.MaxBytesError
		if string(got) != "lar" || !errors.As(err, &mbe) || mbe.Limit != 3 {
			t.Fatalf("limit = %q, %v", got, err)
		}
	}
}

func TestBindBodyCodecsAndReplay(t *testing.T) {
	type payload struct {
		Value string `json:"value" xml:"value"`
	}
	value := strings.Repeat("x", 64<<10)
	for _, tc := range []struct {
		name string
		want any
		new  func() any
	}{
		{"json", &payload{value}, func() any { return new(payload) }},
		{"xml", &payload{value}, func() any { return new(payload) }},
		{"proto", wrapperspb.String(value), func() any { return new(wrapperspb.StringValue) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := coder.GetCoder(tc.name).Marshal(tc.want)
			if err != nil {
				t.Fatal(err)
			}
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(data))
			r.Header.Set("Content-Type", "application/"+tc.name)
			bind := newRouterCoder().BindBody()
			for attempt := 0; attempt < 2; attempt++ {
				out := tc.new()
				if err := bind(r, out); err != nil {
					t.Fatal(err)
				}
				if msg, ok := out.(proto.Message); ok {
					if !proto.Equal(msg, tc.want.(proto.Message)) {
						t.Fatal("protobuf content changed")
					}
				} else if !reflect.DeepEqual(out, tc.want) {
					t.Fatal("decoded content changed")
				}
			}
			replay, err := io.ReadAll(r.Body)
			if err != nil || !bytes.Equal(replay, data) {
				t.Fatal("body replay changed")
			}
		})
	}
}

func TestBindBodyChunked(t *testing.T) {
	r, err := http.ReadRequest(bufio.NewReader(strings.NewReader("POST / HTTP/1.1\r\nHost: example.com\r\nContent-Type: application/json\r\nTransfer-Encoding: chunked\r\n\r\n5\r\n{\"v\":\r\n3\r\n42}\r\n0\r\n\r\n")))
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()
	var out map[string]int
	if err := newRouterCoder().BindBody()(r, &out); err != nil || out["v"] != 42 {
		t.Fatalf("chunked binding = %v, %v", out, err)
	}
}

func TestBindBodyReadAndDecodeErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		body   io.Reader
		replay string
		status int
	}{
		{"empty", strings.NewReader(""), "", 0},
		{"invalid-json", strings.NewReader("{"), "{", http.StatusBadRequest},
		{"trailing-json", strings.NewReader(`{} {}`), `{} {}`, http.StatusBadRequest},
		{"reader-error", io.MultiReader(strings.NewReader(`{}`), iotest.ErrReader(io.ErrUnexpectedEOF)), `{}`, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", tc.body)
			r.ContentLength = 1024
			var out map[string]any
			err := newRouterCoder().BindBody()(r, &out)
			if tc.status == 0 {
				if err != nil || out != nil {
					t.Fatalf("empty binding = %v, %v", out, err)
				}
			} else if err == nil || errorx.Status(err) != tc.status {
				t.Fatalf("binding error = %v", err)
			}
			data, err := io.ReadAll(r.Body)
			if err != nil || string(data) != tc.replay {
				t.Fatalf("error replay = %q, %v", data, err)
			}
		})
	}
}

func TestBindBodyServerLimit(t *testing.T) {
	srv := NewServer(MaxRequestBodySize(16))
	srv.GetRouter().POST("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := NewCtx(w, r)
		var out map[string]any
		err := ctx.Bind(&out)
		if err == nil || errorx.FromError(err).Code != errorx.ErrCodeInvalidReqSys {
			t.Errorf("expected size-limit error, got %v", err)
		}
		ctx.ReturnError(err)
	})
	for _, hint := range []int64{-1, 0, 1, 128, math.MaxInt64} {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"value":"more than sixteen bytes"}`))
		r.ContentLength = hint
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, r)
		if w.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("hint %d: status = %d, body = %s", hint, w.Code, w.Body.String())
		}
	}
}
