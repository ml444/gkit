package consul

import (
	"github.com/hashicorp/consul/api"
	"os"
)

func getConsulAddr() string {
	consulAddr := os.Getenv("CONSUL_ADDR")
	if consulAddr == "" {
		return ""
	}
	return consulAddr
}

func GetClient() (Client, error) {
	cli, err := api.NewClient(&api.Config{
		Address: getConsulAddr(),
	})
	if err != nil {
		return nil, err
	}

	return NewClient(cli), nil
}

type Registration = api.AgentServiceRegistration
