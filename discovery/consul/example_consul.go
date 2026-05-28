package consul

import (
	"context"
	"fmt"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/ml444/gkit/discovery"
)

// ExampleConsulRegistry shows how to use ConsulRegistry
// 展示如何使用ConsulRegistry

func ExampleConsulRegistry() {
	// Create consul client config
	config := consulapi.DefaultConfig()
	config.Address = "localhost:8500" // Default consul agent address

	// Create consul registry with options
	registry, err := NewConsulRegistry(config, 
		WithHealthCheck(true), // Enable health checking
		WithTTL(60),           // Set TTL to 60 seconds
	)
	if err != nil {
		fmt.Printf("Failed to create consul registry: %v\n", err)
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

// ExampleDiscoveryClientWithConsulRegistry shows how to use DiscoveryClient with ConsulRegistry
// 展示如何使用带有ConsulRegistry的DiscoveryClient

func ExampleDiscoveryClientWithConsulRegistry() {
	// Create consul client config
	config := consulapi.DefaultConfig()
	config.Address = "localhost:8500"

	// Create consul registry
	registry, err := NewConsulRegistry(config)
	if err != nil {
		fmt.Printf("Failed to create consul registry: %v\n", err)
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