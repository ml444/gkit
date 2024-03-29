package config

type OptionFunc func(*Config)

// WithFileLoader setIntoStruct the file loader.
func WithFileLoader(loader FileLoader) OptionFunc {
	return func(c *Config) {
		c.fileLoader = loader
	}
}

// WithFilePath setIntoStruct the file path.
func WithFilePath(filePath string) OptionFunc {
	return func(c *Config) {
		c.filePath = filePath
	}
}

func WithIgnoreError(ignoreErr bool) OptionFunc {
	return func(c *Config) {
		c.ignoreErr = ignoreErr
	}
}

// WithEnvKeyPrefix If the environment variable name
// is not specified in the tag of the structure field,
// the prefix plus the field name will be used as the
// environment variable name.
func WithEnvKeyPrefix(prefix string) OptionFunc {
	return func(c *Config) {
		c.envKeyPrefix = prefix
	}
}
