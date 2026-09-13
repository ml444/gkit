package httpx

import (
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"strings"
	"testing"

	"github.com/ml444/gkit/errorx"
	jsoncodec "github.com/ml444/gkit/transport/httpx/coder/json"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestResponseDecoderLimits(t *testing.T) {
	for _, tc := range []struct {
		limit  int64
		body   string
		status int
		large  bool
	}{
		{2, "{}", 200, false}, {1, "{}", 200, true}, {0, "{}", 200, false}, {math.MaxInt64, "{}", 200, false}, {2, "oversize", 400, true},
	} {
		decode, err := NewResponseDecoder(WithDecodeMaxBytes(tc.limit))
		if err != nil {
			t.Fatal(err)
		}
		res := &http.Response{StatusCode: tc.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(tc.body)), ContentLength: -1}
		err = decode(context.Background(), res, &struct{}{})
		var large *ResponseTooLargeError
		if errors.As(err, &large) != tc.large {
			t.Fatalf("limit=%d body=%s err=%v", tc.limit, tc.body, err)
		}
		if tc.large {
			if large.Limit != tc.limit || large.StatusCode != tc.status || errorx.FromError(err).Status != 502 {
				t.Fatal(large)
			}
		} else if err != nil {
			t.Fatal(err)
		}
	}
	for _, opt := range []ResponseDecoderOption{nil, WithDecodeMaxBytes(-1), WithDecodeJSONCoder(nil), WithDecodeJSONCoder((*jsoncodec.ConfiguredCoder)(nil)), WithDecodeJSONCoder(markerCodec{"xml", ""})} {
		if _, err := NewResponseDecoder(opt); err == nil {
			t.Fatal("invalid decoder option accepted")
		}
	}
}

func TestResponseDecoderInstanceAndSkippedBodies(t *testing.T) {
	opts := jsoncodec.DefaultOptions()
	opts.Unmarshal.DiscardUnknown = false
	decode, err := NewResponseDecoder(WithDecodeJSONCoder(jsoncodec.NewCoder(opts)), WithDecodeMaxBytes(64))
	if err != nil {
		t.Fatal(err)
	}
	r := &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"unknown":true}`))}
	if err := decode(context.Background(), r, &descriptorpb.FieldDescriptorProto{}); err == nil {
		t.Fatal("custom JSON decode options ignored")
	}
	for _, status := range []int{204, 205} {
		r.StatusCode = status
		r.Body = io.NopCloser(errorReader{})
		if err := decode(context.Background(), r, &struct{}{}); err != nil {
			t.Fatal(err)
		}
	}
	r.StatusCode = 200
	r.Body = io.NopCloser(errorReader{})
	if err := decode(context.Background(), r, nil); err != nil {
		t.Fatal(err)
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
