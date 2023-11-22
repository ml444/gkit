package json

import "encoding/json"

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) Unmarshal(in []byte, out interface{}) error {
	return json.Unmarshal(in, out)
}
