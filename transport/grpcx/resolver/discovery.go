package resolver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
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
const instanceAttrKey = "gkit.discovery.instance"

var registerOnce sync.Once
var legacyBuilder atomic.Pointer[Builder]
var addrInstanceCache atomic.Value // compatibility-only test cache

// Register registers the first client process-wide. Configure it before dialing.
// Deprecated: use NewBuilder with grpc.WithResolvers to isolate connections.
func Register(dc *discovery.DiscoveryClient) {
	if dc == nil {
		return
	}
	registerOnce.Do(func() {
		b := NewBuilder(dc)
		legacyBuilder.Store(b)
		resolver.Register(b)
	})
}

// Builder binds discovery and feedback state to one client's resolver scope.
// Construct a separate Builder for each client, even when service names match.
type Builder struct {
	dc        *discovery.DiscoveryClient
	mu        sync.RWMutex
	instances map[*discoveryResolver]map[string]discovery.ServiceInstancer
}

type discoveryBuilder = Builder

// NewBuilder creates a connection-scoped resolver builder for grpc.WithResolvers.
func NewBuilder(dc *discovery.DiscoveryClient) *Builder {
	return &Builder{dc: dc, instances: make(map[*discoveryResolver]map[string]discovery.ServiceInstancer)}
}

func (*Builder) Scheme() string { return scheme }

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	if b.dc == nil {
		return nil, errors.New("discovery resolver: DiscoveryClient is required")
	}
	service := parseServiceName(target)
	if service == "" {
		return nil, fmt.Errorf("discovery resolver: empty service name in target %q", target.URL.String())
	}
	r := &discoveryResolver{dc: b.dc, service: service, cc: cc, builder: b, refresh: 30 * time.Second}
	r.start()
	return r, nil
}

// GetInstance returns only an instance belonging to this builder and service.
func (b *Builder) GetInstance(service, addr string) (discovery.ServiceInstancer, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for r, instances := range b.instances {
		if service != "" && r.service != service {
			continue
		}
		if inst, ok := instances[addr]; ok {
			return inst, true
		}
	}
	return nil, false
}

func (b *Builder) setInstances(r *discoveryResolver, instances map[string]discovery.ServiceInstancer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if instances == nil {
		delete(b.instances, r)
		return
	}
	if b.instances == nil {
		b.instances = make(map[*discoveryResolver]map[string]discovery.ServiceInstancer)
	}
	b.instances[r] = instances
}

func InstanceFromAttributes(attrs *attributes.Attributes) discovery.ServiceInstancer {
	if attrs == nil {
		return nil
	}
	inst, _ := attrs.Value(instanceAttrKey).(discovery.ServiceInstancer)
	return inst
}

// GetInstanceByAddr searches only the legacy global registration and test cache.
// Deprecated: use Builder.GetInstance; addresses alone do not identify a service.
func GetInstanceByAddr(addr string) (discovery.ServiceInstancer, bool) {
	if m, ok := addrInstanceCache.Load().(map[string]discovery.ServiceInstancer); ok {
		if inst, ok := m[addr]; ok {
			return inst, true
		}
	}
	if b := legacyBuilder.Load(); b != nil {
		return b.GetInstance("", addr)
	}
	return nil, false
}

// SetInstanceCacheForTest sets a copy of the legacy test cache. It does not affect grpcx clients.
func SetInstanceCacheForTest(m map[string]discovery.ServiceInstancer) {
	cp := make(map[string]discovery.ServiceInstancer, len(m))
	for k, v := range m {
		cp[k] = v
	}
	addrInstanceCache.Store(cp)
}

func parseServiceName(target resolver.Target) string {
	if path := strings.TrimPrefix(target.URL.Path, "/"); path != "" {
		return path
	}
	return strings.TrimPrefix(target.Endpoint(), "/")
}

type discoveryResolver struct {
	dc       *discovery.DiscoveryClient
	service  string
	cc       resolver.ClientConn
	builder  *Builder
	refresh  time.Duration
	stop     chan struct{}
	stopOnce sync.Once
	initOnce sync.Once
	updateMu sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	closed   atomic.Bool
}

func (r *discoveryResolver) init() {
	r.initOnce.Do(func() {
		r.ctx, r.cancel = context.WithCancel(context.Background())
		if r.stop == nil {
			r.stop = make(chan struct{})
		}
	})
}

func (r *discoveryResolver) start() {
	r.init()
	go func() {
		r.update()
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

func (r *discoveryResolver) update() {
	r.init()
	r.updateMu.Lock()
	defer r.updateMu.Unlock()
	if r.closed.Load() {
		return
	}
	ctx, cancel := context.WithTimeout(r.ctx, 10*time.Second)
	defer cancel()
	instances, err := r.dc.GetAllInstances(ctx, r.service)
	if r.closed.Load() {
		return
	}
	if err != nil && !errors.Is(err, discovery.ErrNotFound) {
		log.Errorf("discovery resolver: get instances for %q: %v", r.service, err)
		r.cc.ReportError(err)
		return
	}
	addrs := make([]resolver.Address, 0, len(instances))
	cache := make(map[string]discovery.ServiceInstancer, len(instances))
	for _, inst := range instances {
		addr := net.JoinHostPort(inst.GetAddress(), strconv.Itoa(inst.GetPort()))
		addrs = append(addrs, resolver.Address{Addr: addr, Attributes: attributes.New(instanceAttrKey, inst)})
		cache[addr] = inst
	}
	if r.builder != nil {
		r.builder.setInstances(r, cache)
	}
	if err := r.cc.UpdateState(resolver.State{Addresses: addrs}); err != nil {
		log.Debugf("discovery resolver: update %q: %v", r.service, err)
	}
}

func (r *discoveryResolver) ResolveNow(resolver.ResolveNowOptions) { r.update() }

func (r *discoveryResolver) Close() {
	r.init()
	r.stopOnce.Do(func() { r.closed.Store(true); r.cancel(); close(r.stop) })
	// Wait for an in-flight update before removing its state. No UpdateState is
	// possible after Close returns, and cancellation interrupts discovery I/O.
	r.updateMu.Lock()
	defer r.updateMu.Unlock()
	if r.builder != nil {
		r.builder.setInstances(r, nil)
	}
}
