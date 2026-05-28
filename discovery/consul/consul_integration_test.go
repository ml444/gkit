//go:build integration

package consul

import (
	"context"
	"testing"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ml444/gkit/discovery"
)

func startConsulContainer(t *testing.T) (*consulapi.Client, func()) {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "hashicorp/consul:1.17",
		ExposedPorts: []string{"8500/tcp"},
		Cmd:          []string{"agent", "-dev", "-client=0.0.0.0"},
		WaitingFor:   wait.ForHTTP("/v1/status/leader").WithPort("8500/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start consul: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("host: %v", err)
	}
	port, err := container.MappedPort(ctx, "8500/tcp")
	if err != nil {
		t.Fatalf("port: %v", err)
	}

	client, err := consulapi.NewClient(&consulapi.Config{Address: host + ":" + port.Port()})
	if err != nil {
		t.Fatalf("consul client: %v", err)
	}

	cleanup := func() {
		_ = container.Terminate(ctx)
	}
	return client, cleanup
}

func TestConsulRegistry_Integration(t *testing.T) {
	client, cleanup := startConsulContainer(t)
	defer cleanup()

	reg, err := NewConsulRegistryWithAPI(client.Agent(), client.Health(), WithHealthCheck(false))
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer reg.Close()

	ctx := context.Background()
	ins := &discovery.ServiceInstance{
		ID: "ins-1", Name: "demo", Address: "127.0.0.1", Port: 8080,
	}
	if err := reg.Register(ctx, ins); err != nil {
		t.Fatalf("register: %v", err)
	}

	// Register updates health via agent; query may need a short wait for catalog.
	time.Sleep(500 * time.Millisecond)
	reg.serviceMap.Delete("demo")

	instances, err := reg.GetServiceInstances(ctx, "demo")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(instances) == 0 {
		t.Fatal("expected at least one instance")
	}

	if err := reg.Deregister(ctx, ins); err != nil {
		t.Fatalf("deregister: %v", err)
	}
}
