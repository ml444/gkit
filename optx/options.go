package optx

import (
	"fmt"
	"reflect"
	"strings"
)

func NewOptions(opts ...interface{}) *Options {
	if len(opts)%2 != 0 {
		panic(fmt.Sprintf("invalid number of opts argument %d", len(opts)))
	}
	l := &Options{}
	for i := 0; i < len(opts); i += 2 {
		l.AddOpt(opts[i], opts[i+1])
	}
	return l
}

func (x *Options) IsOptExist(enumKey interface{}) bool {
	int32Key := toInt32(enumKey)

	for _, opt := range x.Options {
		if opt.Kind == int32Key {
			return true
		}
	}

	return false
}

func (x *Options) GetOptValue(enumKey interface{}) (string, bool) {
	int32Key := toInt32(enumKey)
	for _, opt := range x.Options {
		if opt.Kind == int32Key {
			return opt.Value, true
		}
	}

	return "", false
}

func (x *Options) AddOpt(enumKey, val interface{}) *Options {
	int32Key := enumToInt32(enumKey)

	typeOfVal := reflect.TypeOf(val)
	var strVal string
	if val == nil {
		strVal = ""
	} else {
		switch typeOfVal.Kind() {
		case reflect.Slice, reflect.Array:
			vv := reflect.ValueOf(val)
			n := vv.Len()
			var valList []string
			for j := 0; j < n; j++ {
				valList = append(valList, toStr(vv.Index(j).Interface()))
			}
			strVal = strings.Join(valList, ",")
		default:
			strVal = toStr(val)
		}
	}
	x.Options = append(x.Options,
		&Options_Option{Kind: int32Key, Value: strVal})
	return x
}

func enumToInt32(enumKey interface{}) int32 {
	if enum, ok := enumKey.(IEunmType); ok {
		return int32(enum.Number())
	} else {
		return toInt32(enumKey)
	}
}
