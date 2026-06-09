package grpcx

import (
	"context"
	"crypto/tls"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc/credentials/insecure"

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
	if err = srv.RegisterDiscovery(context.Background(), reg,
		RegisterID("custom-id"),
		RegisterVersion("v1"),
		RegisterMetadata(map[string]string{"env": "test"}),
	); err != nil {
		t.Fatalf("RegisterDiscovery: %v", err)
	}
	instances, err := reg.GetServiceInstances(context.Background(), "test-grpc")
	if err != nil {
		t.Fatal(err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(instances))
	}
	if instances[0].GetID() != "custom-id" || instances[0].GetVersion() != "v1" || instances[0].GetMetadata()["env"] != "test" {
		t.Fatalf("unexpected instance: %#v", instances[0])
	}
	if err := srv.DeregisterDiscovery(context.Background(), reg,
		RegisterID("custom-id"),
		RegisterVersion("v1"),
		RegisterMetadata(map[string]string{"env": "test"}),
	); err != nil {
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

func TestNewServerBranches(t *testing.T) {
	if _, err := NewServer(
		Debug(true),
		DisableErrorInterceptor(),
		Credentials(insecure.NewCredentials()),
	); err != nil {
		t.Fatalf("credentials server: %v", err)
	}
	if _, err := NewServer(TLSConfig(&tls.Config{MinVersion: tls.VersionTLS12})); err != nil {
		t.Fatalf("tls server: %v", err)
	}
	t.Setenv("GRPC_XDS_BOOTSTRAP", "/path/that/does/not/exist")
	if _, err := NewServer(EnableXDS()); err == nil {
		t.Fatal("expected xds server error")
	}
}

func TestServerEndpointListenError(t *testing.T) {
	srv, err := NewServer(Network("bad-network"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := srv.Endpoint(); err == nil {
		t.Fatal("expected listen error")
	}
}
