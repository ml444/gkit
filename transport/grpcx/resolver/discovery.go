package resolver

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
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
	currentDC    atomicDiscoveryClient
)

type atomicDiscoveryClient struct {
	mu sync.RWMutex
	dc *discovery.DiscoveryClient
}

func (a *atomicDiscoveryClient) get() *discovery.DiscoveryClient {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.dc
}

func (a *atomicDiscoveryClient) set(dc *discovery.DiscoveryClient) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.dc = dc
}

// Register sets the DiscoveryClient used by the discovery resolver scheme.
// Safe to call before each grpcx.NewClient with discovery enabled.
func Register(dc *discovery.DiscoveryClient) {
	if dc == nil {
		return
	}
	currentDC.set(dc)
	registerOnce.Do(func() {
		resolver.Register(&discoveryBuilder{})
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

type discoveryBuilder struct{}

func (discoveryBuilder) Scheme() string {
	return scheme
}

func (discoveryBuilder) Build(target resolver.Target, cc resolver.ClientConn, _ resolver.BuildOptions) (resolver.Resolver, error) {
	dc := currentDC.get()
	if dc == nil {
		return nil, fmt.Errorf("discovery resolver: DiscoveryClient not registered, call resolver.Register")
	}
	service := parseServiceName(target)
	if service == "" {
		return nil, fmt.Errorf("discovery resolver: empty service name in target %q", target.URL.String())
	}
	r := &discoveryResolver{
		dc:      dc,
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
	dc      *discovery.DiscoveryClient
	service string
	cc      resolver.ClientConn
	refresh time.Duration
	stop    chan struct{}
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

func (r *discoveryResolver) update() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	instances, err := r.dc.GetAllInstances(ctx, r.service)
	if err != nil {
		if err != discovery.ErrNotFound {
			log.Errorf("discovery resolver: get instances for %q: %v", r.service, err)
		}
		_ = r.cc.UpdateState(resolver.State{Addresses: nil})
		return
	}
	addrs := make([]resolver.Address, 0, len(instances))
	for _, inst := range instances {
		addr := net.JoinHostPort(inst.GetAddress(), fmt.Sprintf("%d", inst.GetPort()))
		attrs := attributes.New(instanceAttrKey, inst)
		addrs = append(addrs, resolver.Address{
			Addr:       addr,
			Attributes: attrs,
		})
	}
	_ = r.cc.UpdateState(resolver.State{Addresses: addrs})
}

func (r *discoveryResolver) ResolveNow(resolver.ResolveNowOptions) {
	r.update()
}

func (r *discoveryResolver) Close() {
	close(r.stop)
}
