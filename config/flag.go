package config

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
)

func GetTagValues(tag reflect.StructTag, key string) (bool, []string) {
	tagValue, ok := tag.Lookup(key)
	if !ok {
		return false, nil
	}
	parts := strings.Split(tagValue, ";")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return true, out
}

type tagOptions struct {
	name       string
	defaultStr string
	usage      string
}

// parseStructTagOptions parses tag segments in two forms:
// 1) shorthand name: "PREDICTION_TG_BOT_TOKEN"
// 2) key=value pairs: "name=PREDICTION_TG_BOT_TOKEN; default=abc"
func parseStructTagOptions(tagValues []string) (tagOptions, error) {
	var opts tagOptions
	for _, value := range tagValues {
		key, val, hasKey := strings.Cut(value, "=")
		if !hasKey {
			if opts.name == "" {
				opts.name = value
			}
			continue
		}
		switch strings.TrimSpace(key) {
		case "name":
			opts.name = strings.TrimSpace(val)
		case "default":
			opts.defaultStr = strings.TrimSpace(val)
		case "usage":
			opts.usage = strings.TrimSpace(val)
		default:
			return opts, fmt.Errorf("invalid tag key: %s", strings.TrimSpace(key))
		}
	}
	return opts, nil
}

func parseFieldTagWithFlag(fs *flag.FlagSet, field reflect.StructField, v reflect.Value, mValue *Value, useFlag *bool) error {
	tag := field.Tag
	// Parse the tag of the field
	ok, tagValues := GetTagValues(tag, "flag")
	if ok {
		*useFlag = true
		opts, err := parseStructTagOptions(tagValues)
		if err != nil {
			return err
		}
		name, defaultStr, usage := opts.name, opts.defaultStr, opts.usage
		// name must have a value, because the upper-level logic
		// will determine whether the name is empty or not.
		if name == "" {
			name = field.Name
		}
		mValue.nameInFlag = name
		mValue.kind = originFlag
		if reflect.ValueOf(mValue.value).IsZero() && defaultStr != "" {
			val, err := str2Any(defaultStr, field.Type)
			if err != nil {
				return err
			}
			if val != nil {
				mValue.value = val
				if v.CanSet() {
					v.Set(reflect.ValueOf(val))
				}
			}
		}
		if fs.Lookup(name) != nil {
			return nil
		}
		if v.CanAddr() {
			err = setFlag(fs, name, usage, mValue.value, field.Type, v.Addr())
		} else {
			err = setFlag(fs, name, usage, mValue.value, field.Type, v)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func setFlag(fs *flag.FlagSet, name, usage string, defaultValue interface{}, t reflect.Type, ptr reflect.Value) error {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Convert various types through reflection and assertions
	switch t.Kind() {
	case reflect.Bool:
		fs.BoolVar(ptr.Interface().(*bool), name, defaultValue.(bool), usage)
	case reflect.String:
		fs.StringVar(ptr.Interface().(*string), name, defaultValue.(string), usage)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch t.Kind() {
		case reflect.Int:
			fs.IntVar(ptr.Interface().(*int), name, defaultValue.(int), usage)
		case reflect.Int64:
			fs.Int64Var(ptr.Interface().(*int64), name, defaultValue.(int64), usage)
		default:
			return fmt.Errorf("flag don't support this type[%s], this name of field: [%s]", t.Kind().String(), t.Name())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch t.Kind() {
		case reflect.Uint:
			fs.UintVar(ptr.Interface().(*uint), name, defaultValue.(uint), usage)
		case reflect.Uint64:
			fs.Uint64Var(ptr.Interface().(*uint64), name, defaultValue.(uint64), usage)
		default:
			return fmt.Errorf("flag don't support this type[%s], skip it. FieldName: [%s]", t.Kind().String(), t.Name())
		}
	case reflect.Float32, reflect.Float64:
		switch t.Kind() {
		case reflect.Float64:
			fs.Float64Var(ptr.Interface().(*float64), name, defaultValue.(float64), usage)
		default:
			return fmt.Errorf("flag don't support this type[%s], skip it. FieldName: [%s]", t.Kind().String(), t.Name())
		}
	default:
		return fmt.Errorf("flag don't support this type[%s], skip it. FieldName: [%s]", t.Kind().String(), t.Name())
	}
	return nil
}
