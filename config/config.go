package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
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
	flagSet      *flag.FlagSet
	args         []string
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
		args:         os.Args[1:],
	}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.flagSet == nil {
		cfg.flagSet = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		cfg.flagSet.SetOutput(io.Discard)
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
		parseArgs := filterArgsByDefinedFlags(cfg.flagSet, cfg.args)
		if err = cfg.flagSet.Parse(parseArgs); err != nil && !cfg.ignoreErr {
			return fmt.Errorf("parse flags error: %v", err)
		}
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
		fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		fs.SetOutput(io.Discard)
		fs.StringVar(&fp, c.fileFlag, c.filePath, "Configuration file path")
		parseArgs := filterArgsByDefinedFlags(fs, c.args)
		err = fs.Parse(parseArgs)
		if err != nil && err != flag.ErrHelp {
			return fmt.Errorf("flag parse config filepath error: %w", err)
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

func filterArgsByDefinedFlags(fs *flag.FlagSet, args []string) []string {
	out := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if len(arg) == 0 || arg[0] != '-' {
			continue
		}
		name := arg
		for len(name) > 0 && name[0] == '-' {
			name = name[1:]
		}
		if eq := indexByte(name, '='); eq >= 0 {
			name = name[:eq]
		}
		if name == "" || fs.Lookup(name) == nil {
			continue
		}
		out = append(out, arg)
		if indexByte(arg, '=') < 0 && i+1 < len(args) && len(args[i+1]) > 0 && args[i+1][0] != '-' {
			out = append(out, args[i+1])
			i++
		}
	}
	return out
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func (c *Processor) buildMap(key string, v reflect.Value) (err error) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v = reflect.New(v.Type().Elem()).Elem()
		} else {
			v = v.Elem()
		}
	}

	// 确保是结构体
	if v.Kind() != reflect.Struct {
		return nil
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

		// 嵌套匿名字段（struct），如果已经存在就跳过
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

		vv := v.Field(i)
		// If the field type is a struct, the function calls itself recursively
		// with the updated key and field value.
		if fieldT.Kind() == reflect.Struct {
			// 如果是 nil 指针，初始化可以设置的值
			if isPtrField {
				// vvT := vv.Type()
				// if vvT.Kind() == reflect.Ptr {
				// 	vvT = vvT.Elem()
				// }
				if vv.IsNil() && vv.CanSet() {
					vv.Set(reflect.New(fieldT))
				}
			}

			if err = c.buildMap(k, vv); err != nil && !c.ignoreErr {
				return err
			}
			continue
		} else {
			if !vv.IsValid() {
				continue
			}
			// 如果是指针且 nil，则初始化
			if vv.Kind() == reflect.Ptr {
				if vv.IsNil() && vv.CanSet() {
					vv.Set(reflect.New(vv.Type().Elem()))
				}
				// 解引用
				vv = vv.Elem()
			}

			if vv.CanSet() && vv.IsZero() {
				vv.Set(reflect.Zero(vv.Type()))
			}

			value := vv.Interface()
			if s, ok := value.(string); ok && s != "" {
				value = ReplaceEnvVariables(s)
				if vv.CanSet() {
					vv.Set(reflect.ValueOf(value))
				}
			}
			// []string
			if vv.Kind() == reflect.Slice && vv.Type().Elem().Kind() == reflect.String {
				for j := 0; j < vv.Len(); j++ {
					elem := vv.Index(j)
					orig := elem.String()
					if orig == "" {
						continue
					}
					if vv.CanSet() {
						elem.Set(reflect.ValueOf(ReplaceEnvVariables(orig)))
					}
				}
			}
			mValue := Value{value: value}
			if err = parseFieldTagWithFlag(c.flagSet, rtField, vv, &mValue, &c.useFlag); err != nil && !c.ignoreErr {
				return err
			}
			if err = parseFieldTagWithEnv(rtField, vv, &mValue, k, c.envKeyPrefix); err != nil && !c.ignoreErr {
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
