package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/ml444/gkit/discovery"
)

func TestClient_DiscoveryEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	u := strings.TrimPrefix(ts.URL, "http://")
	host, portStr, ok := strings.Cut(u, ":")
	if !ok {
		t.Fatalf("unexpected server url host: %s", ts.URL)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("invalid port: %v", err)
	}

	reg := discovery.NewDefaultRegistry()
	dc := discovery.NewDiscoveryClient(reg)
	err = reg.Register(context.Background(), &discovery.ServiceInstance{
		ID:      "ins-1",
		Name:    "svc",
		Address: host,
		Port:    port,
	})
	if err != nil {
		t.Fatalf("register instance: %v", err)
	}

	c, err := NewClient(
		WithEndpoint("discovery:///svc"),
		WithDiscovery(dc, ""),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	var out struct {
		OK bool `json:"ok"`
	}
	if err := c.Invoke(context.Background(), http.MethodGet, "/hello", nil, &out); err != nil {
		t.Fatalf("Invoke: %v", err)
	}
	if !out.OK {
		t.Fatalf("expected ok=true, got false")
	}
}

