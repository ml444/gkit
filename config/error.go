package config

import (
	"errors"
	"fmt"
)

var NotFoundValueErr = func(key string) error {
	return errors.New(fmt.Sprintf("not found value with %s", key))
}
