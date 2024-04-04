package pluck

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

const DisablePluckHeader = "Gkit-Disable-Pluck"

func CopyBodyFromRequest(req *http.Request) (data []byte, err error) {
	data, err = io.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(data))
	return
}

func ConvertAnyToHeader(v interface{}, isBytes bool) (header http.Header, err error) {
	header = make(http.Header)
	if isBytes {
		header.Set("Content-Type", "application/octet-stream")
	}
	if v == nil {
		return
	}
	_v := reflect.ValueOf(v)
	if _v.Kind() == reflect.Ptr {
		_v = _v.Elem()
	}
	if _v.Kind() == reflect.Map {
		if vv, ok := _v.Interface().(map[string][]string); ok {
			for k, v := range vv {
				header[k] = v
			}
			return
		} else if vv, ok := _v.Interface().(map[string]string); ok {
			for k, v := range vv {
				header[k] = []string{v}
			}
			return
		}
		err = errors.New("v must be a map[string][]string or map[string]string")
		return
	}
	if _v.Kind() == reflect.Struct {
		for i := 0; i < _v.NumField(); i++ {
			field := _v.Field(i)
			key := ProtoFieldNameToKey(_v.Type().Field(i).Name)
			if !field.IsValid() || !field.CanInterface() {
				continue
			}
			if field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.String {
				if field.Len() == 0 {
					continue
				}
				header[key] = field.Interface().([]string)
			} else if field.Kind() == reflect.String {
				if field.String() == "" {
					continue
				}
				header[key] = []string{field.String()}
			} else {
				err = fmt.Errorf("the field type of struct [%T] is not supported", v)
				return
			}
		}
	}
	return

}

func ExtractHeader(header http.Header, v interface{}) error {
	_v := reflect.ValueOf(v)
	if _v.Kind() == reflect.Ptr {
		_v = _v.Elem()
	}
	if _v.IsNil() {
		switch _v.Kind() {
		case reflect.Ptr:
			_v.Set(reflect.New(_v.Type().Elem()))
		case reflect.Map:
			_v.Set(reflect.MakeMap(_v.Type()))
		default:
			return errors.New("v must be a pointer of *struct or map")
		}
	}

	if _v.Kind() == reflect.Ptr {
		_v = _v.Elem()
	}
	switch _v.Kind() {
	case reflect.Map:
		if vv, ok := _v.Interface().(map[string][]string); ok {
			for k, v := range header {
				vv[k] = v
			}
			return nil
		} else if vv, ok := _v.Interface().(map[string]string); ok {
			//if vv == nil {
			//	vv = map[string]string{}
			//}
			for k, v := range header {
				vv[k] = strings.Join(v, ",")
			}
			return nil
		}
		return errors.New("v must be a map[string][]string or map[string]string")
	case reflect.Struct:
		// If v is a structure type, map the request header to the structure field
		//for i := 0; i < _v.NumField(); i++ {
		//	field := _v.Field(i)
		for key, values := range header {
			field := _v.FieldByName(KeyToProtoField(key))
			if !field.IsValid() {
				continue
			}
			if field.CanSet() {
				err := setValueToField(field, values)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("the field [%s] of struct [%T] is not found or cannot be set", key, v)
			}
		}

	default:
		return errors.New("v must be a pointer to a map or struct")
	}
	return nil
}

func SetResponseHeaders(w http.ResponseWriter, v interface{}) error {
	_v := reflect.ValueOf(v)
	if v == nil {
		_v.Elem().Set(reflect.New(reflect.TypeOf(v)))
	}
	if (_v.Kind() != reflect.Ptr && _v.Kind() != reflect.Map) || (_v.Kind() == reflect.Ptr && _v.Elem().Kind() != reflect.Struct) {
		return errors.New("h must be a pointer of struct, or map")
	}
	if _v.IsNil() {
		return errors.New("h must not be nil")
	}
	if _v.Kind() == reflect.Ptr {
		_v = _v.Elem()
	}
	switch _v.Kind() {
	case reflect.Map:
		if vv, ok := v.(map[string][]string); ok {
			for k, v := range vv {
				w.Header().Set(k, strings.Join(v, ","))
			}
			return nil
		} else if vv, ok := v.(map[string]string); ok {
			for k, v := range vv {
				w.Header().Set(k, v)
			}
			return nil
		}
		return errors.New("h must be a map[string][]string or map[string]string")
	case reflect.Struct:
		for i := 0; i < _v.NumField(); i++ {
			field := _v.Field(i)
			key := ProtoFieldNameToKey(_v.Type().Field(i).Name)
			if !field.CanSet() {
				return fmt.Errorf("the field [%s] of struct [%T] is not found or cannot be set", key, v)
			}
			if field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.String {
				if field.Len() == 0 {
					continue
				}
				w.Header().Set(key, strings.Join(field.Interface().([]string), ","))
			} else if field.Kind() == reflect.String {
				if field.String() == "" {
					continue
				}
				w.Header().Set(key, field.String())
			} else {
				return fmt.Errorf("the field type of struct [%T] is not supported", v)
			}
		}
	default:
		return errors.New("h must be a pointer to a map or struct")
	}
	return nil
}

func setValueToField(field reflect.Value, values []string) error {
	if len(values) == 0 {
		return errors.New("values is empty")
	}
	switch field.Kind() {
	case reflect.Slice: // reflect.Array:
		elementT := field.Type().Elem()
		vList := reflect.MakeSlice(field.Type(), 0, len(values))
		for _, v := range values {
			elementV := reflect.New(elementT)
			err := setValueToField(elementV.Elem(), []string{v})
			if err != nil {
				return err
			}
			vList = reflect.Append(vList, elementV.Elem())
		}
		field.Set(vList)
	case reflect.String:
		field.SetString(strings.Join(values, ","))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v := values[0]
		int64V, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(int64V)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v := values[0]
		uint64V, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uint64V)
	case reflect.Float32, reflect.Float64:
		v := values[0]
		float64V, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		field.SetFloat(float64V)
	case reflect.Bool:
		v := values[0]
		boolV, err := strconv.ParseBool(v)
		if err != nil {
			return err
		}
		field.SetBool(boolV)
	case reflect.Pointer:
		// field = field.Elem()
		fallthrough
	case reflect.Struct, reflect.Map:
		v := values[0]
		res := reflect.New(field.Type())
		err := json.Unmarshal([]byte(v), res.Interface())
		if err != nil {
			return err
		}
		field.Set(res.Elem())
	default:
		return fmt.Errorf("cannot support type of %s", field.Kind().String())
	}
	return nil
}

func toTitle(s string) string {
	if s == "" {
		return s
	}
	if d := s[0]; d >= 'a' && d <= 'z' {
		return string(d-32) + s[1:]
	}
	return s
}

// KeyToProtoField Convert key with dash '-' to Protobuf field name
func KeyToProtoField(key string) string {
	// Convert dashes in keys to camelCase notation
	parts := strings.Split(key, "-")
	for i, part := range parts {
		//if i == 0 {
		//	continue
		//}
		parts[i] = toTitle(part)
	}
	return strings.Join(parts, "")
}

// ProtoFieldNameToKey Convert Protobuf message field names to HTTP request header field names
func ProtoFieldNameToKey(protoField string) string {
	// Convert camel case naming to dash `-` connection naming
	var builder strings.Builder
	for i, r := range protoField {
		if i > 1 && 'A' <= r && r <= 'Z' {
			builder.WriteRune('-')
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

// ProtoMessageFieldToHeaders 将 Protobuf 消息字段转换为 HTTP 请求头
func ProtoMessageFieldToHeaders() {

}
