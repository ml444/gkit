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

// IsLocalEnv returns true if current env is local or not set to it.
func IsLocalEnv() bool {
	if name := GetCurrentEnvName(); name == "" || name == LocalEnvName {
		return true
	}
	return false
}

// IsEnv returns true if current env is envName.
func IsEnv(envName string) bool {
	if GetCurrentEnvName() == envName {
		return true
	}
	return false
}

// GteEnv returns true if current env is greater than envName.
func GteEnv(envName string) bool {
	if GetEnvValue(GetCurrentEnvName()) >= GetEnvValue(envName) {
		return true
	}
	return false
}

// LteEnv returns true if current env is less than envName.
func LteEnv(envName string) bool {
	if GetEnvValue(GetCurrentEnvName()) <= GetEnvValue(envName) {
		return true
	}
	return false
}
