package core

import "os"

const (
	KeyCurrentEnvName = "GKIT_CURRENT_ENV_NAME"
)

const (
	LocalEnv = iota + 1
	DevelopmentEnv
	TestingEnv
	ProductEnv
)

const (
	LocalEnvName       = "local"
	DevelopmentEnvName = "development"
	TestingEnvName     = "testing"
	ProductEnvName     = "product"
)

func GetCurrentEnv() int {
	envName := os.Getenv(KeyCurrentEnvName)
	switch envName {
	case ProductEnvName:
		return ProductEnv
	case TestingEnvName:
		return TestingEnv
	case DevelopmentEnvName:
		return DevelopmentEnv
	case LocalEnvName:
		return LocalEnv
	default:
		return 0
	}
}

func IsLocalEnv() bool {
	if GetCurrentEnv() <= 1 {
		return true
	}
	return false
}
