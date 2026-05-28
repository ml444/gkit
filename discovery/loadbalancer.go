package discovery

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// LoadBalancerType represents the type of load balancer
// 负载均衡器类型
type LoadBalancerType string

const (
	// RandomLoadBalancer selects a random instance
	RandomLoadBalancer LoadBalancerType = "random"
	// RoundRobinLoadBalancer selects instances in round-robin order
	RoundRobinLoadBalancer LoadBalancerType = "round_robin"
	// LeastConnectionsLoadBalancer selects the instance with the least connections
	LeastConnectionsLoadBalancer LoadBalancerType = "least_connections"
)

// LoadBalancer is the interface for load balancing
// 负载均衡器接口
type LoadBalancer interface {
	// Select selects a service instance based on load balancing algorithm
	// 根据负载均衡算法选择一个服务实例
	Select(ctx context.Context, instances []ServiceInstancer) (ServiceInstancer, error)
	// Update updates the load balancer state when a request is completed
	// 请求完成后更新负载均衡器状态
	Update(ctx context.Context, instance ServiceInstancer, success bool)
}

// RandomLoadBalancerImpl implements random load balancing
// 随机负载均衡实现
type RandomLoadBalancerImpl struct {
	random *rand.Rand
	mu     sync.Mutex
}

// NewRandomLoadBalancer creates a new RandomLoadBalancer
func NewRandomLoadBalancer() *RandomLoadBalancerImpl {
	return NewRandomLoadBalancerWithRand(rand.New(rand.NewSource(time.Now().UnixNano())))
}

// NewRandomLoadBalancerWithRand creates a RandomLoadBalancer with a custom rand source.
func NewRandomLoadBalancerWithRand(r *rand.Rand) *RandomLoadBalancerImpl {
	return &RandomLoadBalancerImpl{
		random: r,
	}
}

// Select selects a random instance
func (lb *RandomLoadBalancerImpl) Select(ctx context.Context, instances []ServiceInstancer) (ServiceInstancer, error) {
	if len(instances) == 0 {
		return nil, ErrNotFound
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()
	return instances[lb.random.Intn(len(instances))], nil
}

// Update updates the load balancer state
func (lb *RandomLoadBalancerImpl) Update(ctx context.Context, instance ServiceInstancer, success bool) {
	// Random load balancer doesn't need to update state
}

// RoundRobinLoadBalancerImpl implements round-robin load balancing
// 轮询负载均衡实现
type RoundRobinLoadBalancerImpl struct {
	counter uint64
}

// NewRoundRobinLoadBalancer creates a new RoundRobinLoadBalancer
func NewRoundRobinLoadBalancer() *RoundRobinLoadBalancerImpl {
	return &RoundRobinLoadBalancerImpl{
		counter: 0,
	}
}

// Select selects the next instance in round-robin order
func (lb *RoundRobinLoadBalancerImpl) Select(ctx context.Context, instances []ServiceInstancer) (ServiceInstancer, error) {
	if len(instances) == 0 {
		return nil, ErrNotFound
	}

	// Atomically increment the counter and get the index
	counter := atomic.AddUint64(&lb.counter, 1) - 1
	index := counter % uint64(len(instances))
	return instances[index], nil
}

// Update updates the load balancer state
func (lb *RoundRobinLoadBalancerImpl) Update(ctx context.Context, instance ServiceInstancer, success bool) {
	// Round-robin load balancer doesn't need to update state
}

// LeastConnectionsLoadBalancerImpl implements least connections load balancing
// 最小连接数负载均衡实现
type LeastConnectionsLoadBalancerImpl struct {
	connections map[string]int32
	mu          sync.Mutex
}

// NewLeastConnectionsLoadBalancer creates a new LeastConnectionsLoadBalancer
func NewLeastConnectionsLoadBalancer() *LeastConnectionsLoadBalancerImpl {
	return &LeastConnectionsLoadBalancerImpl{
		connections: make(map[string]int32),
	}
}

// Select selects the instance with the least connections
func (lb *LeastConnectionsLoadBalancerImpl) Select(ctx context.Context, instances []ServiceInstancer) (ServiceInstancer, error) {
	if len(instances) == 0 {
		return nil, ErrNotFound
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	active := make(map[string]struct{}, len(instances))
	for _, instance := range instances {
		active[instance.GetID()] = struct{}{}
		if _, exists := lb.connections[instance.GetID()]; !exists {
			lb.connections[instance.GetID()] = 0
		}
	}
	for id := range lb.connections {
		if _, ok := active[id]; !ok {
			delete(lb.connections, id)
		}
	}

	var selected ServiceInstancer
	var minConn int32 = -1

	for _, instance := range instances {
		conn := lb.connections[instance.GetID()]
		if minConn == -1 || conn < minConn {
			minConn = conn
			selected = instance
		}
	}

	// Increment connection count for selected instance
	lb.connections[selected.GetID()]++

	return selected, nil
}

// Update updates the load balancer state
func (lb *LeastConnectionsLoadBalancerImpl) Update(ctx context.Context, instance ServiceInstancer, success bool) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Decrement connection count
	if conn, exists := lb.connections[instance.GetID()]; exists && conn > 0 {
		lb.connections[instance.GetID()]--
	}
}

// NewLoadBalancer creates a new load balancer based on the given type
// 根据类型创建负载均衡器
func NewLoadBalancer(lbType LoadBalancerType) (LoadBalancer, error) {
	switch lbType {
	case RandomLoadBalancer:
		return NewRandomLoadBalancer(), nil
	case RoundRobinLoadBalancer:
		return NewRoundRobinLoadBalancer(), nil
	case LeastConnectionsLoadBalancer:
		return NewLeastConnectionsLoadBalancer(), nil
	default:
		return nil, errors.New("unsupported load balancer type")
	}
}
