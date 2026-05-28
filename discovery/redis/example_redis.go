package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ml444/gkit/discovery"
)

// ExampleRedisRegistry shows how to use RedisRegistry
// 展示如何使用RedisRegistry

func ExampleRedisRegistry() {
	// Create redis registry with options
	registry, err := NewRedisRegistry(
		WithTTL(60),        // Set TTL to 60 seconds
		WithPrefix("gkit:services"), // Set custom prefix
	)
	if err != nil {
		fmt.Printf("Failed to create redis registry: %v\n", err)
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

// ExampleRedisRegistryWithCustomClient shows how to use RedisRegistry with custom client
// 展示如何使用带有自定义客户端的RedisRegistry

func ExampleRedisRegistryWithCustomClient() {
	// Create custom Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Create redis registry with custom client
	registry, err := NewRedisRegistryWithClient(client, WithTTL(30))
	if err != nil {
		fmt.Printf("Failed to create redis registry with custom client: %v\n", err)
		return
	}
	defer registry.Close()

	// Use the registry as usual
	// ...
}

// ExampleDiscoveryClientWithRedisRegistry shows how to use DiscoveryClient with RedisRegistry
// 展示如何使用带有RedisRegistry的DiscoveryClient

func ExampleDiscoveryClientWithRedisRegistry() {

	// Create redis registry
	registry, err := NewRedisRegistry()
	if err != nil {
		fmt.Printf("Failed to create redis registry: %v\n", err)
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