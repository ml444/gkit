package k8s

import (
	"context"
	"fmt"

	"github.com/ml444/gkit/discovery"
)

// ExampleK8sRegistry demonstrates how to use the Kubernetes service registry
// 演示如何使用Kubernetes服务注册发现
func ExampleK8sRegistry() {
	// 创建一个K8sRegistry实例（假设在Kubernetes集群内运行）
	registry, err := NewK8sRegistry(
		WithNamespace("default"),
	)
	if err != nil {
		fmt.Printf("Failed to create K8sRegistry: %v\n", err)
		return
	}
	defer registry.Close()

	// 使用Kubernetes服务注册发现
	instances, err := registry.GetServiceInstances(context.Background(), "example-service")
	if err != nil && err != discovery.ErrNotFound {
		fmt.Printf("Failed to get service instances: %v\n", err)
		return
	}

	fmt.Printf("Found %d instances of service 'example-service'\n", len(instances))
	for _, instance := range instances {
		fmt.Printf("Instance: %s, Address: %s:%d\n", instance.GetID(), instance.GetAddress(), instance.GetPort())
	}
}

// ExampleDiscoveryClientWithK8sRegistry demonstrates how to use DiscoveryClient with K8sRegistry
// 演示如何结合K8sRegistry使用DiscoveryClient
func ExampleDiscoveryClientWithK8sRegistry() {
	// 创建一个K8sRegistry实例
	registry, err := NewK8sRegistry(
		WithNamespace("default"),
	)
	if err != nil {
		fmt.Printf("Failed to create K8sRegistry: %v\n", err)
		return
	}
	defer registry.Close()

	// 创建一个使用K8sRegistry的DiscoveryClient
	client := discovery.NewDiscoveryClient(
		registry,
		discovery.WithLoadBalancer(discovery.NewRandomLoadBalancer()),
	)
	if err != nil {
		fmt.Printf("Failed to create DiscoveryClient: %v\n", err)
		return
	}

	// 使用DiscoveryClient获取服务实例
	instance, err := client.GetServiceInstance(context.Background(), "example-service")
	if err != nil {
		fmt.Printf("Failed to get service instance: %v\n", err)
		return
	}

	fmt.Printf("Selected instance: %s, Address: %s:%d\n", instance.GetID(), instance.GetAddress(), instance.GetPort())
}