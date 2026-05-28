package consul

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/ml444/gkit/discovery"
)

type fakeAgent struct {
	services map[string]*consulapi.AgentServiceRegistration
}

func (f *fakeAgent) ServiceRegister(reg *consulapi.AgentServiceRegistration) error {
	f.services[reg.ID] = reg
	return nil
}

func (f *fakeAgent) ServiceDeregister(id string) error {
	delete(f.services, id)
	return nil
}

func (f *fakeAgent) UpdateTTL(string, string, string) error { return nil }

type fakeHealth struct {
	entries map[string][]*consulapi.ServiceEntry
}

func (f *fakeHealth) Service(name, tag string, passingOnly bool, _ *consulapi.QueryOptions) ([]*consulapi.ServiceEntry, *consulapi.QueryMeta, error) {
	entries := f.entries[name]
	if len(entries) == 0 {
		return nil, nil, nil
	}
	return entries, nil, nil
}

func TestConsulRegistry_RegisterGetDeregister(t *testing.T) {
	agent := &fakeAgent{services: make(map[string]*consulapi.AgentServiceRegistration)}
	health := &fakeHealth{entries: make(map[string][]*consulapi.ServiceEntry)}

	reg, err := NewConsulRegistryWithAPI(agent, health, WithHealthCheck(false))
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer reg.Close()

	ctx := context.Background()
	meta, _ := json.Marshal(map[string]string{"env": "test"})
	health.entries["user"] = []*consulapi.ServiceEntry{{
		Service: &consulapi.AgentService{
			ID:      "ins-1",
			Service: "user",
			Address: "10.0.0.1",
			Port:    8080,
			Meta: map[string]string{
				"metadata": string(meta),
			},
		},
	}}

	ins := &discovery.ServiceInstance{
		ID: "ins-1", Name: "user", Address: "10.0.0.1", Port: 8080,
		Metadata: map[string]string{"env": "test"},
	}
	if err := reg.Register(ctx, ins); err != nil {
		t.Fatalf("register: %v", err)
	}

	instances, err := reg.GetServiceInstances(ctx, "user")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(instances) != 1 || instances[0].GetMetadata()["env"] != "test" {
		t.Fatalf("unexpected instances: %+v", instances)
	}

	instances[0] = &discovery.ServiceInstance{ID: "mutated"}
	cached, err := reg.GetServiceInstances(ctx, "user")
	if err != nil {
		t.Fatalf("get cached: %v", err)
	}
	if cached[0].GetID() != "ins-1" {
		t.Fatal("cache should not be mutated by caller")
	}

	if err := reg.Deregister(ctx, ins); err != nil {
		t.Fatalf("deregister: %v", err)
	}
	delete(health.entries, "user")
	time.Sleep(20 * time.Millisecond)
	_, err = reg.GetServiceInstances(ctx, "user")
	if !errors.Is(err, discovery.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestConsulRegistry_DeregisterInvalidatesCache(t *testing.T) {
	agent := &fakeAgent{services: make(map[string]*consulapi.AgentServiceRegistration)}
	health := &fakeHealth{entries: map[string][]*consulapi.ServiceEntry{
		"svc": {{
			Service: &consulapi.AgentService{ID: "1", Service: "svc", Address: "127.0.0.1", Port: 8080},
		}},
	}}

	reg, err := NewConsulRegistryWithAPI(agent, health, WithHealthCheck(false))
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer reg.Close()

	ctx := context.Background()
	if _, err := reg.GetServiceInstances(ctx, "svc"); err != nil {
		t.Fatalf("prime cache: %v", err)
	}

	health.entries["svc"] = nil
	if err := reg.Deregister(ctx, &discovery.ServiceInstance{ID: "1", Name: "svc"}); err != nil {
		t.Fatalf("deregister: %v", err)
	}

	time.Sleep(20 * time.Millisecond)
	_, err = reg.GetServiceInstances(ctx, "svc")
	if !errors.Is(err, discovery.ErrNotFound) {
		t.Fatalf("expected cache invalidation after deregister, got %v", err)
	}
}
