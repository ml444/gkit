package nacos

import (
	"context"
	"fmt"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"

	"github.com/ml444/gkit/discovery"
)

// ExampleNacosRegistry shows how to use NacosRegistry
// 展示如何使用NacosRegistry

func ExampleNacosRegistry() {
	// Create nacos server configs
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: "localhost",
			Port:   8848, // Default nacos server port
		},
	}

	// Create nacos client config
	clientConfig := constant.ClientConfig{
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogLevel:            "warn",
	}

	// Create nacos registry with options
	registry, err := NewNacosRegistry(serverConfigs, clientConfig, 
		WithTimeout(3*time.Second), // Set operation timeout
	)
	if err != nil {
		fmt.Printf("Failed to create nacos registry: %v\n", err)
		return
	}
	defer registry.Close()

	// Create service instance
	instance := &discovery.ServiceInstance{
		ID:       "service-1",
		Name:     "userService",
		Address:  "127.0.0.1",
		Port:     8080,
		Metadata: map[string]string{"version": "v1", "environment": "dev"},
	}

	// Register service
	ctx := context.Background()
	err = registry.Register(ctx, instance)
	if err != nil {
		fmt.Printf("Failed to register service: %v\n", err)
		return
	}
	fmt.Println("Service registered successfully")

	// Get service instances
	instances, err := registry.GetServiceInstances(ctx, "userService")
	if err != nil {
		fmt.Printf("Failed to get service instances: %v\n", err)
	} else {
		fmt.Printf("Found %d instances\n", len(instances))
		for i, ins := range instances {
			fmt.Printf("Instance %d: %s:%d\n", i, ins.GetAddress(), ins.GetPort())
		}
	}

	// Subscribe to service changes
	err = registry.Subscribe("userService", func(instances []discovery.ServiceInstancer) {
		fmt.Printf("Service instances changed, now have %d instances\n", len(instances))
	})
	if err != nil {
		fmt.Printf("Failed to subscribe to service: %v\n", err)
	}

	// Wait for a while
	time.Sleep(2 * time.Second)

	// Deregister service
	err = registry.Deregister(ctx, instance)
	if err != nil {
		fmt.Printf("Failed to deregister service: %v\n", err)
	} else {
		fmt.Println("Service deregistered successfully")
	}
}

// ExampleDiscoveryClientWithNacosRegistry shows how to use DiscoveryClient with NacosRegistry
// 展示如何使用带有NacosRegistry的DiscoveryClient

func ExampleDiscoveryClientWithNacosRegistry() {
	// Create nacos server configs
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: "localhost",
			Port:   8848,
		},
	}

	// Create nacos client config
	clientConfig := constant.ClientConfig{
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
	}

	// Create nacos registry
	registry, err := NewNacosRegistry(serverConfigs, clientConfig)
	if err != nil {
		fmt.Printf("Failed to create nacos registry: %v\n", err)
		return
	}

	// Create discovery client with the registry
	client := discovery.NewDiscoveryClient(registry,
		discovery.WithRefreshInterval(30*time.Second),
	)
	defer client.Close()

	// Get a service instance using load balancing
	instance, err := client.GetServiceInstance(context.Background(), "userService")
	if err != nil {
		fmt.Printf("Failed to get service instance: %v\n", err)
	} else if instance != nil {
		fmt.Printf("Selected instance: %s:%d\n", instance.GetAddress(), instance.GetPort())
	}
}