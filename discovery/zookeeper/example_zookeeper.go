package zookeeper

import (
	"context"
	"fmt"
	"time"

	"github.com/ml444/gkit/discovery"
)

// ExampleZookeeperRegistry shows how to use ZookeeperRegistry
// 展示如何使用ZookeeperRegistry

func ExampleZookeeperRegistry() {
	// Create zookeeper registry with options
	registry, err := NewZookeeperRegistry([]string{"localhost:2181"}, 
		WithTTL(60),        // Set TTL to 60 seconds
		WithBasePath("/gkit/services"), // Set custom base path
	)
	if err != nil {
		fmt.Printf("Failed to create zookeeper registry: %v\n", err)
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

// ExampleDiscoveryClientWithZookeeperRegistry shows how to use DiscoveryClient with ZookeeperRegistry
// 展示如何使用带有ZookeeperRegistry的DiscoveryClient

func ExampleDiscoveryClientWithZookeeperRegistry() {

	// Create zookeeper registry
	registry, err := NewZookeeperRegistry([]string{"localhost:2181"})
	if err != nil {
		fmt.Printf("Failed to create zookeeper registry: %v\n", err)
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