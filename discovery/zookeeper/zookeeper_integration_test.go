//go:build integration

package zookeeper

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ml444/gkit/discovery"
)

func TestZookeeperRegistry_Integration(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "zookeeper:3.9",
		ExposedPorts: []string{"2181/tcp"},
		WaitingFor:   wait.ForListeningPort("2181/tcp").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start zookeeper: %v", err)
	}
	defer container.Terminate(ctx)

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("host: %v", err)
	}
	port, err := container.MappedPort(ctx, "2181/tcp")
	if err != nil {
		t.Fatalf("port: %v", err)
	}

	reg, err := NewZookeeperRegistry([]string{host + ":" + port.Port()})
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer reg.Close()

	ins := &discovery.ServiceInstance{
		ID: "1", Name: "svc", Address: "127.0.0.1", Port: 8080,
	}
	if err := reg.Register(context.Background(), ins); err != nil {
		t.Fatalf("register: %v", err)
	}

	reg.mu.Lock()
	delete(reg.services, "svc")
	reg.mu.Unlock()

	instances, err := reg.GetServiceInstances(context.Background(), "svc")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(instances))
	}

	if err := reg.Deregister(context.Background(), ins); err != nil {
		t.Fatalf("deregister: %v", err)
	}
	_, err = reg.GetServiceInstances(context.Background(), "svc")
	if err == nil {
		t.Fatal("expected error after deregister")
	}
}
