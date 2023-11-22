package toml

import "github.com/BurntSushi/toml"

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) Unmarshal(in []byte, out interface{}) error {
	err := toml.Unmarshal(in, out)
	if err != nil {
		return err
	}
	return nil
}
