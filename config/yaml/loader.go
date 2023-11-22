package yaml

import "gopkg.in/yaml.v3"

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) Unmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}
