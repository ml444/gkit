package consul

import (
	stdconsul "github.com/hashicorp/consul/api"

	"github.com/ml444/gkit/log"
)

// Registrar registers service instance liveness information to Consul.
type Registrar struct {
	client       Client
	registration *stdconsul.AgentServiceRegistration
	logger       log.Logger
}

// NewRegistrar returns a Consul Registrar acting on the provided catalog
// registration.
func NewRegistrar(client Client, r *stdconsul.AgentServiceRegistration, logger log.Logger) *Registrar {
	return &Registrar{
		client:       client,
		registration: r,
		logger:       logger,
	}
}

// Register implements sd.Registrar interface.
func (p *Registrar) Register() {
	if err := p.client.Register(p.registration); err != nil {
		p.logger.Info("err:", err)
	} else {
		p.logger.Info("action:", "register")
	}
}

// Deregister implements sd.Registrar interface.
func (p *Registrar) Deregister() {
	if err := p.client.Deregister(p.registration); err != nil {
		p.logger.Info("err:", err)
	} else {
		p.logger.Info("action:", "deregister")
	}
}
