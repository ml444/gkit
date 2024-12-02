package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
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

type Processor struct {
	v            interface{}
	m            map[string]*Value
	fileLoader   FileLoader
	filePath     string
	fileFlag     string
	envKeyPrefix string
	ignoreErr    bool
	useFlag      bool
}

// InitConfig passes in a structure pointer.
// Recursively traverse all fields and construct a map.
// Get the value from the environment variable and overwrite the value in the map if it exists.
func InitConfig(v interface{}, opts ...OptionFunc) error {
	if v == nil {
		panic("InitConfig: v is nil")
	}

	var err error
	cfg := &Processor{
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
		return err
	}

	// Get all fields in the structure (including nested fields) to build a map
	err = cfg.buildMap("", reflect.ValueOf(v))
	if err != nil && !cfg.ignoreErr {
		return fmt.Errorf("build map error: %v", err)
	}
	if cfg.useFlag {
		flag.Parse()
	}
	return nil
}

func (c *Processor) loadFromFile() (err error) {
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

func (c *Processor) buildMap(key string, v reflect.Value) (err error) {
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
				for j := 0; j < vv.Len(); j++ {
					s := vv.Index(j).String()
					if s == "" {
						continue
					}
					vv.Index(j).Set(reflect.ValueOf(ReplaceEnvVariables(s)))
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

// Walk Pass in a config structure and a key-value processing function,
// traverse the structure, and pass the field names and values to the function for processing
func Walk(c any, fn func(k string, v any) error) error {
	val := reflect.ValueOf(c)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return errors.New("provided value is not a struct")
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		if !field.IsExported() {
			continue
		}
		fieldValue := val.Field(i)

		// Call handler function
		err := fn(field.Name, fieldValue.Interface())
		if err != nil {
			return err
		}
	}

	return nil
}
