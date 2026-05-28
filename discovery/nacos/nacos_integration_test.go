//go:build integration

package nacos

import (
	"context"
	"testing"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ml444/gkit/discovery"
)

func TestNacosRegistry_Integration(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "nacos/nacos-server:v2.3.2",
		ExposedPorts: []string{"8848/tcp"},
		Env: map[string]string{
			"MODE":                       "standalone",
			"NACOS_AUTH_ENABLE":          "false",
			"JVM_XMS":                    "256m",
			"JVM_XMX":                    "256m",
		},
		WaitingFor: wait.ForHTTP("/nacos/v1/console/health/readiness").WithPort("8848/tcp").WithStartupTimeout(3 * time.Minute),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start nacos: %v", err)
	}
	defer container.Terminate(ctx)

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("host: %v", err)
	}
	port, err := container.MappedPort(ctx, "8848/tcp")
	if err != nil {
		t.Fatalf("port: %v", err)
	}

	reg, err := NewNacosRegistry(
		[]constant.ServerConfig{{IpAddr: host, Port: uint64(port.Int())}},
		constant.ClientConfig{
			NamespaceId:         "",
			TimeoutMs:           5000,
			NotLoadCacheAtStart: true,
			LogDir:              t.TempDir(),
			CacheDir:            t.TempDir(),
		},
		WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer reg.Close()

	ins := &discovery.ServiceInstance{
		ID: "1", Name: "demo", Address: "127.0.0.1", Port: 8080,
	}
	if err := reg.Register(context.Background(), ins); err != nil {
		t.Fatalf("register: %v", err)
	}

	time.Sleep(time.Second)
	reg.serviceMap.Delete("demo")

	instances, err := reg.GetServiceInstances(context.Background(), "demo")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(instances) == 0 {
		t.Fatal("expected instances")
	}
}
