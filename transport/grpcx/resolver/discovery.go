package resolver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"

	"github.com/ml444/gkit/discovery"
	"github.com/ml444/gkit/log"
)

const scheme = "discovery"

// instanceAttrKey is stored in resolver.Address.Attributes for load-balancer feedback.
const instanceAttrKey = "gkit.discovery.instance"

var (
	registerOnce sync.Once
)

// Register sets the DiscoveryClient used by the discovery resolver scheme.
// Safe to call before each grpcx.NewClient with discovery enabled.
func Register(dc *discovery.DiscoveryClient) {
	if dc == nil {
		return
	}
	// currentDC.set(dc)
	registerOnce.Do(func() {
		resolver.Register(&discoveryBuilder{dc: dc})
	})
}

// InstanceFromAttributes returns the ServiceInstancer attached to an address.
func InstanceFromAttributes(attrs *attributes.Attributes) discovery.ServiceInstancer {
	if attrs == nil {
		return nil
	}
	v := attrs.Value(instanceAttrKey)
	if v == nil {
		return nil
	}
	inst, _ := v.(discovery.ServiceInstancer)
	return inst
}

type discoveryBuilder struct {
	dc *discovery.DiscoveryClient
}

func (b discoveryBuilder) Scheme() string {
	return scheme
}

func (b discoveryBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	if b.dc == nil {
		return nil, errors.New("discovery resolver: DiscoveryClient not registered, call resolver.Register")
	}
	service := parseServiceName(target)
	if service == "" {
		return nil, fmt.Errorf("discovery resolver: empty service name in target %q", target.URL.String())
	}
	r := &discoveryResolver{
		dc:      b.dc,
		service: service,
		cc:      cc,
		refresh: 30 * time.Second,
		stop:    make(chan struct{}),
	}
	r.start()
	return r, nil
}

func parseServiceName(target resolver.Target) string {
	path := strings.TrimPrefix(target.URL.Path, "/")
	if path != "" {
		return path
	}
	return strings.TrimPrefix(target.Endpoint(), "/")
}

type discoveryResolver struct {
	dc       *discovery.DiscoveryClient
	service  string
	cc       resolver.ClientConn
	refresh  time.Duration
	stop     chan struct{}
	stopOnce sync.Once
}

func (r *discoveryResolver) start() {
	r.update()
	go func() {
		ticker := time.NewTicker(r.refresh)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				r.update()
			case <-r.stop:
				return
			}
		}
	}()
}

// 🌟 1. 定义包级缓存（原子替换，无锁且无内存泄露风险）
var addrInstanceCache atomic.Value // 存储类型为 map[string]discovery.ServiceInstancer

// 🌟 2. 提供 O(1) 查询方法供 client.go 的拦截器调用
func GetInstanceByAddr(addr string) (discovery.ServiceInstancer, bool) {
	m, _ := addrInstanceCache.Load().(map[string]discovery.ServiceInstancer)
	if m == nil {
		return nil, false
	}
	inst, ok := m[addr]
	return inst, ok
}

// SetInstanceCacheForTest 仅用于单元测试，手动注入实例缓存
func SetInstanceCacheForTest(m map[string]discovery.ServiceInstancer) {
	addrInstanceCache.Store(m)
}

func (r *discoveryResolver) update() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	instances, err := r.dc.GetAllInstances(ctx, r.service)
	if err != nil {
		if errors.Is(err, discovery.ErrNotFound) {
			// Service genuinely has no instances; reflect the empty state.
			_ = r.cc.UpdateState(resolver.State{Addresses: []resolver.Address{}})
			return
		}
		// Transient error (registry blip): keep the last-good address set and
		// only report the error instead of dropping all backends.
		log.Errorf("discovery resolver: get instances for %q: %v", r.service, err)
		r.cc.ReportError(err)
		return
	}
	addrs := make([]resolver.Address, 0, len(instances))
	// 🌟 3. 构建一个新的 map，用于原子替换
	newCache := make(map[string]discovery.ServiceInstancer, len(instances))
	for _, inst := range instances {
		addr := net.JoinHostPort(inst.GetAddress(), fmt.Sprintf("%d", inst.GetPort()))
		attrs := attributes.New(instanceAttrKey, inst)
		addrs = append(addrs, resolver.Address{
			Addr:       addr,
			Attributes: attrs,
		})
		// 🌟 4. 将实例写入新的 map 中
		newCache[addr] = inst
	}
	// 🌟 5. 原子替换缓存，热路径读端（client.go）完全无锁
	addrInstanceCache.Store(newCache)
	_ = r.cc.UpdateState(resolver.State{Addresses: addrs})
}

func (r *discoveryResolver) ResolveNow(resolver.ResolveNowOptions) {
	r.update()
}

func (r *discoveryResolver) Close() {
	r.stopOnce.Do(func() {
		close(r.stop)
	})
}
