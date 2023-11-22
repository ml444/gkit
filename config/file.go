package config

import (
	"os"
)

type FileLoader interface {
	Unmarshal(in []byte, out interface{}) error
}

// LoadFile load the configuration through the configuration file,
// supporting multiple formats, such as: yaml, json, toml, ini.
// Note: This method will not overwrite existing configurations in the
// command line and environment variables, only other configurations.
func (c *Config) LoadFile() error {
	yamlFile, err := os.ReadFile(c.filePath)
	if err != nil {
		return err
	}
	err = c.fileLoader.Unmarshal(yamlFile, c.v)
	if err != nil {
		return err
	}
	return nil
}
