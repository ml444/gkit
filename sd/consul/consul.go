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
		Scheme:  "https",
		Address: getConsulAddr(),
		Token:   "571ce07c-6771-9091-afbf-9958e715fba9",
	})
	if err != nil {
		return nil, err
	}

	return NewClient(cli), nil
}

type Registration = api.AgentServiceRegistration
