package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
)

const delimiter = "__"

// OriginKind label the source of the value
type OriginKind int

const (
	originFlag OriginKind = iota + 1
	originEnv
	originFile
	originInner
)

type Value struct {
	kind       OriginKind
	nameInFlag string
	nameInEnv  string
	nameInFile string
	value      interface{}
}

func (v Value) Kind() OriginKind {
	return v.kind
}
func (v Value) FlagName() string {
	return v.nameInFlag
}
func (v Value) EnvName() string {
	return v.nameInEnv
}
func (v Value) FieldName() string {
	return v.nameInFile
}
func (v Value) Value() interface{} {
	return v.value
}

type Config struct {
	v            interface{}
	m            map[string]*Value
	fileLoader   FileLoader
	filePath     string
	fileFlag     string
	envKeyPrefix string
	ignoreErr    bool
	useFlag      bool
}

// InitConfig passes in a structure pointer and returns a Config object.
// Recursively traverse all fields and construct a map.
// Get the value from the environment variable and overwrite the value in the map if it exists.
func InitConfig(v interface{}, opts ...OptionFunc) (*Config, error) {
	if v == nil {
		panic("InitConfig: v is nil")
	}

	var err error
	cfg := &Config{
		v:            v,
		m:            make(map[string]*Value),
		ignoreErr:    false,
		envKeyPrefix: "",
		fileFlag:     "",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	err = cfg.loadFromFile()
	if err != nil && !cfg.ignoreErr {
		return nil, err
	}

	// Get all fields in the structure (including nested fields) to build a map
	err = cfg.buildMap("", reflect.ValueOf(v))
	if err != nil && !cfg.ignoreErr {
		return nil, fmt.Errorf("build map error: %v", err)
	}
	if cfg.useFlag {
		if cfg.fileFlag != "" {
			flag.StringVar(&cfg.filePath, cfg.fileFlag, cfg.filePath, "Configuration file path")
		}
		flag.Parse()
	}
	return cfg, nil
}

func (c *Config) loadFromFile() (err error) {
	if c.filePath == "" && c.fileFlag == "" {
		return nil
	}
	if c.fileFlag != "" {
		// Get the configuration file path through fileFlag
		var fp string
		fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		fs.StringVar(&fp, c.fileFlag, c.filePath, "Configuration file path")
		err = fs.Parse(os.Args[1:])
		if err != nil {
			return fmt.Errorf("flag parse config filepath error: %v", err)
		}
		if fp != "" {
			c.filePath = fp
		}
	}
	err = c.LoadFile()
	if err != nil && !c.ignoreErr {
		return fmt.Errorf("load config file error: %v", err)
	}
	return nil
}

func (c *Config) buildMap(key string, v reflect.Value) (err error) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v = reflect.Zero(v.Type().Elem())
		} else {
			v = v.Elem()
		}
	}

	for i := 0; i < v.NumField(); i++ {
		rtField := v.Type().Field(i)
		if !rtField.IsExported() {
			continue
		}

		k := rtField.Name
		if key != "" {
			k = key + delimiter + k
		}

		vv := v.Field(i)

		if rtField.Anonymous {
			if _, ok := c.m[k]; ok {
				continue
			}
		}

		isPtrField := false
		fieldT := rtField.Type
		if fieldT.Kind() == reflect.Ptr {
			fieldT = fieldT.Elem()
			isPtrField = true
		}

		// If the field type is a struct, the function calls itself recursively
		// with the updated key and field value.
		if fieldT.Kind() == reflect.Struct {
			if isPtrField && vv.IsNil() {
				vvT := vv.Type()
				if vvT.Kind() == reflect.Ptr {
					vvT = vvT.Elem()
				}
				vv.Set(reflect.New(vvT))
			}
			err = c.buildMap(k, vv)
			if err != nil && !c.ignoreErr {
				return err
			}
		} else {
			var value interface{}
			if vv.IsZero() {
				value = reflect.Zero(vv.Type()).Interface()
				if vv.CanSet() {
					vv.Set(reflect.ValueOf(value))
				}
			} else {
				value = vv.Interface()
			}
			if s, ok := value.(string); ok && s != "" {
				value = ReplaceEnvVariables(s)
				vv.Set(reflect.ValueOf(value))
			}
			// []string
			if vv.Kind() == reflect.Slice && vv.Type().Elem().Kind() == reflect.String {
				for i := 0; i < vv.Len(); i++ {
					s := vv.Index(i).String()
					if s == "" {
						continue
					}
					vv.Index(i).Set(reflect.ValueOf(ReplaceEnvVariables(s)))
				}
			}
			mValue := Value{value: value}
			if err = parseFieldTagWithFlag(rtField, vv, &mValue, &c.useFlag); err != nil && !c.ignoreErr {
				return err
			}
			if err = parseFieldTagWithEnv(rtField, vv, &mValue); err != nil && !c.ignoreErr {
				return err
			}
			c.m[k] = &mValue
		}
	}
	return nil
}

// Get can obtain the value in the structure
func (c *Config) Get(parts ...string) (value interface{}) {
	// If no parts are provided, return the value stored in c.v
	if keysLen := len(parts); keysLen == 0 {
		return c.v
	} else if keysLen == 1 {
		// If only one part is provided, check if it exists in c.m
		k := parts[0]
		if v, ok := c.m[k]; ok {
			return v.value
		}
		// If not found in c.m, split the key using delimiter
		parts = strings.Split(k, delimiter)
	} else {
		if v, ok := c.m[strings.Join(parts, delimiter)]; ok {
			return v.value
		}
	}
	return value
}

// Set to setIntoStruct the value in the structure
func (c *Config) Set(key string, v interface{}) error {
	// set to Config
	err := c.setIntoStruct(strings.Split(key, delimiter), v)
	if err != nil {
		return err
	}

	old, ok := c.m[key]
	if !ok {
		return NotFoundValueErr(key)
	}
	old.value = v
	c.m[key] = old

	return nil
}

func (c *Config) setIntoStruct(parts []string, v interface{}) error {
	vV := reflect.ValueOf(c.v)
	for _, key := range parts {
		if vV.Kind() == reflect.Ptr {
			vV = vV.Elem()
		}

		vV = vV.FieldByName(key)
		if !vV.IsValid() {
			return errors.New("invalid key")
		}

		cfgT := vV.Type()
		if cfgT.Kind() == reflect.Ptr {
			cfgT = cfgT.Elem()
		}

		if cfgT.Kind() == reflect.Struct {
			continue
		}
		if vStr, ok := v.(string); ok {
			value, err := str2Any(vStr, vV.Type())
			if err != nil && !c.ignoreErr {
				return err
			}
			v = value
		}
		vV.Set(reflect.ValueOf(v))
		break
	}
	return nil
}

// SetAndChangeEnv setIntoStruct the value in the environment variable
func (c *Config) SetAndChangeEnv(key string, v string) error {

	err := c.Set(key, v)
	if err != nil {
		return err
	}

	// set to env
	mValue, ok := c.m[key]
	if !ok {
		return NotFoundValueErr(key)
	}
	if mValue.nameInEnv != "" {
		return os.Setenv(mValue.nameInEnv, v)
	}
	return os.Setenv(key, v)
}

func (c *Config) Walk(fn func(k string, v *Value) error) error {
	if c.m == nil {
		return nil
	}
	for k, v := range c.m {
		err := fn(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
