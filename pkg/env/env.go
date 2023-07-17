package env

import "os"

const (
	KeyCurrentEnvName = "GKIT_ENV_NAME"
)

const (
	LocalEnv = iota + 1
	DevelopEnv
	TestingEnv
	ProductEnv
)

const (
	LocalEnvName   = "local"
	DevelopEnvName = "develop"
	TestingEnvName = "testing"
	ProductEnvName = "product"
)

func GetCurrentEnvName() string {
	return os.Getenv(KeyCurrentEnvName)
}

func GetEnvValue(envName string) int {
	switch envName {
	case ProductEnvName:
		return ProductEnv
	case TestingEnvName:
		return TestingEnv
	case DevelopEnvName:
		return DevelopEnv
	case LocalEnvName:
		return LocalEnv
	default:
		return 0
	}
}

func IsLocalEnv() bool {
	if GetCurrentEnvName() == LocalEnvName {
		return true
	}
	return false
}

func IsEnv(envName string) bool {
	if GetCurrentEnvName() == envName {
		return true
	}
	return false
}

func GteEnv(envName string) bool {
	if GetEnvValue(GetCurrentEnvName()) >= GetEnvValue(envName) {
		return true
	}
	return false
}
