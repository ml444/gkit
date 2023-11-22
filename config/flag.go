package config

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
)

func GetTagValues(tag reflect.StructTag, key string) (bool, []string) {
	tagValue, ok := tag.Lookup(key)
	return ok, strings.Split(tagValue, ";")
}
func GetTagValue(tag reflect.StructTag, key string) (bool, string) {
	tagValue, ok := tag.Lookup(key)
	return ok, tagValue
}

func slice2map(values []string) map[string]string {
	for _, value := range values {
		if !strings.Contains(value, "=") {
			continue
		}
		sList := strings.Split(value, "=")
		if len(sList) != 2 {
			panic("")
		}
	}
	return nil
}

func parseFieldTagWithFlag(field reflect.StructField, v reflect.Value, mValue *Value, useFlag *bool) error {
	tag := field.Tag
	// Parse the tag of the field
	ok, tagValues := GetTagValues(tag, "flag")
	if ok {
		*useFlag = true
		name, defaultStr, usage := "", "", ""
		for _, value := range tagValues {
			sList := strings.SplitN(value, "=", 2)
			if len(sList) != 2 {
				return fmt.Errorf("invalid value: %s", value)
			}
			switch sList[0] {
			case "name":
				name = sList[1]
			case "default":
				defaultStr = sList[1]
			case "usage":
				usage = sList[1]
			}
		}
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
		var err error
		if v.CanAddr() {
			err = setFlag(name, usage, mValue.value, field.Type, v.Addr())
		} else {
			err = setFlag(name, usage, mValue.value, field.Type, v)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func setFlag(name, usage string, defaultValue interface{}, t reflect.Type, ptr reflect.Value) error {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Convert various types through reflection and assertions
	switch t.Kind() {
	case reflect.Bool:
		flag.BoolVar(ptr.Interface().(*bool), name, defaultValue.(bool), usage)
	case reflect.String:
		flag.StringVar(ptr.Interface().(*string), name, defaultValue.(string), usage)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch t.Kind() {
		case reflect.Int:
			flag.IntVar(ptr.Interface().(*int), name, defaultValue.(int), usage)
		case reflect.Int64:
			flag.Int64Var(ptr.Interface().(*int64), name, defaultValue.(int64), usage)
		default:
			return fmt.Errorf("flag don't support this type[%s], this name of field: [%s]", t.Kind().String(), t.Name())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch t.Kind() {
		case reflect.Uint:
			flag.UintVar(ptr.Interface().(*uint), name, defaultValue.(uint), usage)
		case reflect.Uint64:
			flag.Uint64Var(ptr.Interface().(*uint64), name, defaultValue.(uint64), usage)
		default:
			return fmt.Errorf("flag don't support this type[%s], skip it. FieldName: [%s]", t.Kind().String(), t.Name())
		}
	case reflect.Float32, reflect.Float64:
		switch t.Kind() {
		case reflect.Float64:
			flag.Float64Var(ptr.Interface().(*float64), name, defaultValue.(float64), usage)
		default:
			return fmt.Errorf("flag don't support this type[%s], skip it. FieldName: [%s]", t.Kind().String(), t.Name())
		}
	default:
		return fmt.Errorf("flag don't support this type[%s], skip it. FieldName: [%s]", t.Kind().String(), t.Name())
	}
	return nil
}
