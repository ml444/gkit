package config

import (
	"os"
	"errors"
)

type FileLoader interface {
	Unmarshal(in []byte, out interface{}) error
}

// LoadFile load the configuration through the configuration file,
// supporting multiple formats, such as: yaml, json, toml, ini.
// Note: This method will not overwrite existing configurations in the
// command line and environment variables, only other configurations.
func (c *Config) LoadFile() error {
	if c.filePath == "" {
		return nil
	}
	if c.fileLoader == nil {
		return errors.New("config file loader is required")
		//switch {
		//case strings.HasSuffix(c.filePath, ".yml") || strings.HasSuffix(c.filePath, ".yaml"):
		//	c.fileLoader = yaml.NewLoader()
		//case strings.HasSuffix(c.filePath, ".ini") || strings.HasSuffix(c.filePath, ".cfg"):
		//	c.fileLoader = ini.NewLoader()
		//case strings.HasSuffix(c.filePath, ".json"):
		//	c.fileLoader = json.NewLoader()
		//case strings.HasSuffix(c.filePath, ".toml"):
		//	c.fileLoader = toml.NewLoader()
		//default:
		//	c.fileLoader = yaml.NewLoader()
		//}
	}
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
