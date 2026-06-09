package httpx

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ml444/gkit/discovery"
)

func TestClientInvokeBranches(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/ok":
			_, _ = w.Write([]byte(`{"ok":true}`))
		case "/bad":
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":400,"code":1,"message":"bad"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	c, err := NewClient(WithEndpoint(strings.TrimPrefix(ts.URL, "http://")))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.Invoke(context.Background(), http.MethodGet, "/ok", nil, &out, OnResponse(func(res *http.Response) error {
		if res.StatusCode != http.StatusOK {
			t.Fatalf("status = %d", res.StatusCode)
		}
		return nil
	})); err != nil || !out.OK {
		t.Fatalf("invoke ok = %#v %v", out, err)
	}
	if err := c.Invoke(context.Background(), http.MethodGet, "/bad", nil, &out); err == nil {
		t.Fatal("expected decoded error")
	}
	if err := c.Invoke(context.Background(), http.MethodGet, "/ok", nil, &out, OnResponse(func(*http.Response) error {
		return errors.New("hook")
	})); err == nil {
		t.Fatal("expected on response error")
	}
}

func TestClientInvokeEncoderAndDoErrors(t *testing.T) {
	c, err := NewClient(
		WithEndpoint("example.com"),
		WithRequestEncoder(func(context.Context, string, interface{}) ([]byte, error) {
			return nil, errors.New("encode")
		}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Invoke(context.Background(), http.MethodPost, "/x", struct{}{}, nil); err == nil {
		t.Fatal("expected encoder error")
	}

	c, err = NewClient(WithEndpoint("discovery:///"))
	if err != nil {
		t.Fatal(err)
	}
	c.discovery = discovery.NewDiscoveryClient(discovery.NewDefaultRegistry())
	req := httptest.NewRequest(http.MethodGet, "http://discovery/x", nil)
	if _, _, err := c.Do(req); err == nil {
		t.Fatal("expected empty discovery service error")
	}

	if _, err := NewClient(WithEndpoint("http://[::1")); err == nil {
		t.Fatal("expected invalid endpoint error")
	}
}

func TestClientUpdateDiscoveryStatus(t *testing.T) {
	reg := discovery.NewDefaultRegistry()
	dc := discovery.NewDiscoveryClient(reg)
	client := &Client{discovery: dc}
	inst := &discovery.ServiceInstance{ID: "i1", Name: "svc", Address: "127.0.0.1", Port: 80}
	client.updateDiscoveryStatus(context.Background(), true)
	client.updateDiscoveryStatus(context.WithValue(context.Background(), discoveryInstanceKey{}, "bad"), true)
	client.updateDiscoveryStatus(context.WithValue(context.Background(), discoveryInstanceKey{}, inst), true)
}

func TestClientCloseNilAndNonClosingTransport(t *testing.T) {
	var nilClient *Client
	if err := nilClient.Close(); err != nil {
		t.Fatalf("nil close: %v", err)
	}
	c := &Client{transport: roundTripFunc(func(*http.Request) (*http.Response, error) { return nil, nil })}
	if err := c.Close(); err != nil {
		t.Fatalf("close non-closing transport: %v", err)
	}
}
