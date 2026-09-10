package storage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/transport"
	"github.com/ml444/gkit/transport/httpx"
)

type uploadEcho struct{ StorageHTTPServer }

func (uploadEcho) UploadV0(_ context.Context, in *UploadReq) (*UploadRsp, error) {
	return &UploadRsp{Url: in.GetFileInfo().GetFileName(), Size: uint32(len(in.GetFileData()))}, nil
}

func TestGeneratedClientAndHandlerRoundTrip(t *testing.T) {
	calls := 0
	mw := func(next middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, in interface{}) (interface{}, error) {
			calls++
			tr, _ := transport.FromContext(ctx)
			tr.Out().Append("Set-Cookie", "a=1", "b=2")
			return next(ctx, in)
		}
	}
	srv := httpx.NewServer(httpx.SetMiddlewares(mw))
	RegisterStorageHTTPServerWithPrefix(srv, "/api", uploadEcho{})
	ts := httptest.NewServer(srv)
	defer ts.Close()
	client, err := httpx.NewClient(httpx.WithEndpoint(ts.URL + "/api"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	generated := NewStorageHTTPClient(client)
	out, err := generated.UploadV0(context.Background(), &UploadReq{FileInfo: &FileInfo{FileName: "file?/中文"}, FileData: []byte{0xfb, 0xff}}, httpx.OnResponse(func(r *http.Response) error {
		if len(r.Cookies()) != 2 {
			t.Errorf("multi-value headers lost: %v", r.Header)
		}
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if out.GetUrl() != "file?/中文" || out.GetSize() != 2 || calls != 1 {
		t.Fatalf("reply=%v middleware calls=%d", out, calls)
	}
}
