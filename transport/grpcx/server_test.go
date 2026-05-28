package grpcx

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ml444/gkit/discovery"
)

func TestServer_EndpointAndRegisterDiscovery(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	srv, err := NewServer(
		Listener(ln),
		EnableHealth(),
		Name("test-grpc"),
	)
	if err != nil {
		t.Fatal(err)
	}
	ep, err := srv.Endpoint()
	if err != nil {
		t.Fatalf("Endpoint: %v", err)
	}
	if ep == "" {
		t.Fatal("expected endpoint")
	}

	reg := discovery.NewDefaultRegistry()
	if err := srv.RegisterDiscovery(context.Background(), reg); err != nil {
		t.Fatalf("RegisterDiscovery: %v", err)
	}
	instances, err := reg.GetServiceInstances(context.Background(), "test-grpc")
	if err != nil {
		t.Fatal(err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(instances))
	}
	if err := srv.DeregisterDiscovery(context.Background(), reg); err != nil {
		t.Fatalf("DeregisterDiscovery: %v", err)
	}
}

func TestServer_StopWithContext(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv, err := NewServer(Listener(ln))
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		_ = srv.Start()
		close(done)
	}()
	time.Sleep(50 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = srv.Stop(ctx)
	<-done
}
