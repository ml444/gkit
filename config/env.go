package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// loadEnv Get the value from the environment variable and overwrite the value in the map
func loadEnv(m map[string]*Value, ignoreErr bool, prefix string) error {
	for k, v := range m {
		envKey := prefix + k
		if prefix != "" {
			envKey = prefix + strings.ToUpper(strings.ReplaceAll(k, delimiter, "_"))
		}
		if v.nameInEnv != "" {
			envKey = v.nameInEnv
		}
		if strValue := os.Getenv(envKey); strValue != "" {
			vT := reflect.TypeOf(v.value)
			if vT.Kind() == reflect.Ptr {
				vT = vT.Elem()
			}

			// Convert various types through reflection and assertions
			value, err := str2Any(strValue, vT)
			if err != nil && !ignoreErr {
				return err
			}
			if value != nil {
				v.value = value
				v.kind = originEnv
			}
		}
	}

	return nil
}

func parseFieldTagWithEnv(field reflect.StructField, v reflect.Value, mValue *Value, key, prefix string) error {
	tag := field.Tag
	// Parse the tag of the field
	ok, tagValues := GetTagValues(tag, "env")
	name, defaultStr := "", ""
	if ok {
		for _, value := range tagValues {
			sList := strings.SplitN(value, "=", 2)
			if len(sList) != 2 {
				return fmt.Errorf("invalid value: %s, need format with 'key=value'", value)
			}
			switch sList[0] {
			case "name":
				name = sList[1]
			case "default":
				defaultStr = sList[1]
			}
		}
	}
	if name == "" && prefix != "" {
		name = prefix + strings.ToUpper(strings.ReplaceAll(key, delimiter, "_"))
	}
	if name == "" {
		return nil
	}
	var val interface{}
	var err error
	if strValue := os.Getenv(name); strValue != "" {
		val, err = str2Any(strValue, field.Type)
		if err != nil {
			return err
		}
	} else if ok && reflect.ValueOf(mValue.value).IsZero() && defaultStr != "" {
		val, err = str2Any(defaultStr, field.Type)
		if err != nil {
			return err
		}
	}
	if val != nil {
		mValue.value = val
		rv := reflect.ValueOf(val)
		if rv.Type().AssignableTo(v.Type()) {
			v.Set(rv)
		} else if rv.Type().ConvertibleTo(v.Type()) {
			v.Set(rv.Convert(v.Type()))
		}
		mValue.kind = originEnv
	}
	mValue.nameInEnv = name
	return nil
}
