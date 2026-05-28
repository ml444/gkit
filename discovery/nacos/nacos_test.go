package nacos

import (
	"context"
	"errors"
	"testing"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"

	"github.com/ml444/gkit/discovery"
)

type fakeNamingClient struct {
	instances map[string][]model.Instance
}

func (f *fakeNamingClient) RegisterInstance(param vo.RegisterInstanceParam) (bool, error) {
	f.instances[param.ServiceName] = append(f.instances[param.ServiceName], model.Instance{
		Ip:       param.Ip,
		Port:     param.Port,
		Metadata: param.Metadata,
	})
	return true, nil
}

func (f *fakeNamingClient) DeregisterInstance(param vo.DeregisterInstanceParam) (bool, error) {
	list := f.instances[param.ServiceName]
	remaining := make([]model.Instance, 0, len(list))
	for _, ins := range list {
		if ins.Ip != param.Ip || ins.Port != param.Port {
			remaining = append(remaining, ins)
		}
	}
	f.instances[param.ServiceName] = remaining
	return true, nil
}

func (f *fakeNamingClient) SelectInstances(param vo.SelectInstancesParam) ([]model.Instance, error) {
	return f.instances[param.ServiceName], nil
}

func (f *fakeNamingClient) Subscribe(param *vo.SubscribeParam) error {
	if param.SubscribeCallback != nil {
		param.SubscribeCallback(f.instances[param.ServiceName], nil)
	}
	return nil
}

func TestNacosRegistry_RegisterGetDeregister(t *testing.T) {
	client := &fakeNamingClient{instances: make(map[string][]model.Instance)}
	reg, err := newNacosRegistry(client)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer reg.Close()

	ctx := context.Background()
	ins := &discovery.ServiceInstance{
		ID: "1", Name: "order", Address: "127.0.0.1", Port: 8080,
		Metadata: map[string]string{"zone": "a"},
	}
	if err := reg.Register(ctx, ins); err != nil {
		t.Fatalf("register: %v", err)
	}

	instances, err := reg.GetServiceInstances(ctx, "order")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(instances) != 1 || instances[0].GetID() != "127.0.0.1:8080" {
		t.Fatalf("unexpected instances: %+v", instances)
	}

	if err := reg.Deregister(ctx, ins); err != nil {
		t.Fatalf("deregister: %v", err)
	}
	_, err = reg.GetServiceInstances(ctx, "order")
	if !errors.Is(err, discovery.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestNacosRegistry_SubscribeUpdatesCache(t *testing.T) {
	client := &fakeNamingClient{instances: map[string][]model.Instance{
		"pay": {{Ip: "10.0.0.2", Port: 9000}},
	}}
	reg, err := newNacosRegistry(client)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer reg.Close()

	if err := reg.Subscribe("pay", nil); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	instances, err := reg.GetServiceInstances(context.Background(), "pay")
	if err != nil {
		t.Fatalf("get cached: %v", err)
	}
	if len(instances) != 1 || instances[0].GetAddress() != "10.0.0.2" {
		t.Fatalf("unexpected instances: %+v", instances)
	}
}
