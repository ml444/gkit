package env

import (
    "os"
    "fmt"
)

func ValidateEnvs(keys ...string) error {
	var missedKeys []string
	for _, key := range keys {
		if os.Getenv(key) == "" {
			missedKeys = append(missedKeys, key)
		}
	}
	if len(missedKeys) > 0 {
		return fmt.Errorf("missing environment variables: %s", missedKeys)
	}
	return nil
}
