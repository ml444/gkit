package ini

import "gopkg.in/ini.v1"

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) Unmarshal(in []byte, out interface{}) error {
	cfg, err := ini.Load(in)
	if err != nil {
		return err
	}
	err = cfg.MapTo(out)
	if err != nil {
		return err
	}
	return nil
}
