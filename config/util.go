package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/ml444/gkit/log"
)

// str2Any Only the type of the base is processed and converted
func str2Any(strValue string, t reflect.Type) (v interface{}, err error) {
	var isPtr bool
	if strValue == "" {
		return reflect.Zero(t).Interface(), nil
	}
	if t.Kind() == reflect.Ptr {
		isPtr = true
		t = t.Elem()
	}

	// Convert various types through reflection and assertions
	switch t.Kind() {
	case reflect.Bool:
		vBool, err := strconv.ParseBool(strValue)
		if err != nil {
			log.Error("loadEnv err: ", err)
			return nil, err
		}
		if isPtr {
			return &vBool, nil
		}
		return vBool, nil
	case reflect.String:
		if isPtr {
			return &strValue, nil
		}
		return strValue, nil
	case reflect.Uintptr:
		vUint64, err := strconv.ParseUint(strValue, 10, 64)
		if err != nil {
			log.Error("loadEnv err: ", err)
			return nil, err
		}
		return uintptr(vUint64), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		vInt64, err := strconv.ParseInt(strValue, 10, 64)
		if err != nil {
			return nil, err
		}
		switch t.Kind() {
		case reflect.Int:
			resV := int(vInt64)
			if isPtr {
				return &resV, nil
			}
			return resV, nil
		case reflect.Int8:
			resV := int8(vInt64)
			if isPtr {
				return &resV, nil
			}
			return resV, nil
		case reflect.Int16:
			resV := int16(vInt64)
			if isPtr {
				return &resV, nil
			}
			return resV, nil
		case reflect.Int32:
			resV := int32(vInt64)
			if isPtr {
				return &resV, nil
			}
			return resV, nil
		case reflect.Int64:
			if isPtr {
				return &vInt64, nil
			}
			return vInt64, nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		vUint64, err := strconv.ParseUint(strValue, 10, 64)
		if err != nil {
			log.Error("loadEnv err: ", err)
			return nil, err
		}
		switch t.Kind() {
		case reflect.Uint:
			resV := uint(vUint64)
			if isPtr {
				return &resV, nil
			}
			return resV, nil
		case reflect.Uint8:
			resV := uint8(vUint64)
			if isPtr {
				return &resV, nil
			}
			return resV, nil
		case reflect.Uint16:
			resV := uint16(vUint64)
			if isPtr {
				return &resV, nil
			}
			return resV, nil
		case reflect.Uint32:
			resV := uint32(vUint64)
			if isPtr {
				return &resV, nil
			}
			return resV, nil
		case reflect.Uint64:
			if isPtr {
				return &vUint64, nil
			}
			return vUint64, nil
		}
	case reflect.Float32, reflect.Float64:
		vFloat64, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			log.Error("loadEnv err: ", err)
			return nil, err
		}
		switch t.Kind() {
		case reflect.Float32:
			resV := float32(vFloat64)
			if isPtr {
				return &resV, nil
			}
			return resV, nil
		case reflect.Float64:
			if isPtr {
				return &vFloat64, nil
			}
			return vFloat64, nil
		}

	case reflect.Array, reflect.Slice:
		if t.Kind() == reflect.Array {
			return nil, fmt.Errorf("the field [%s] does not use array types", t.Name())
		}
		elemT := t.Elem()
		sList := strings.Split(strValue, ",")
		var vList = reflect.MakeSlice(t, 0, len(sList))

		for _, s := range sList {
			ptr := reflect.New(elemT)
			vv, err := str2Any(s, elemT)
			if err != nil {
				return nil, err
			}
			if vv == nil {
				continue
			}
			ptr.Elem().Set(reflect.ValueOf(vv))
			vList = reflect.Append(vList, ptr.Elem())
		}
		return vList.Interface(), nil
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("the key of map must be string")
		}
		elemT := t.Elem()
		if elemT.Kind() == reflect.Ptr {
			elemT = elemT.Elem()
		}
		sList := strings.Split(strValue, ",")
		vMap := reflect.MakeMap(t)
		for _, s := range sList {
			kv := strings.Split(s, ":")
			if len(kv) != 2 {
				return nil, fmt.Errorf("the value of map must be key:value")
			}
			ptr := reflect.New(elemT)
			vv, err := str2Any(kv[1], elemT)
			if err != nil {
				return nil, err
			}
			if vv == nil {
				continue
			}
			ptr.Elem().Set(reflect.ValueOf(vv))
			vMap.SetMapIndex(reflect.ValueOf(kv[0]), ptr.Elem())
		}
		return vMap.Interface(), nil
	case reflect.Interface:
		return strValue, nil
	default:
		return nil, fmt.Errorf("not support type of %s", t.Kind().String())
	}

	return reflect.Zero(t).Interface(), nil
}
