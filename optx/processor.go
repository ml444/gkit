package optx

import (
	"fmt"
	"reflect"
	"strconv"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type IEunmType interface {
	String() string
	Number() protoreflect.EnumNumber
}

type Processor struct {
	keyMap     map[int32]string
	handlerMap map[string]Handler
	skipZero   bool
}

func NewProcessor() *Processor {
	return &Processor{
		handlerMap: make(map[string]Handler),
		keyMap:     map[int32]string{},
	}
}

func (p *Processor) SkipZero() *Processor {
	p.skipZero = true
	return p
}

// SetHandler method: Only a few types are accepted: string/IEnumType/String()/Integer
func (p *Processor) SetHandler(key interface{}, h Handler) {
	switch keyV := key.(type) {
	case string:
		p.handlerMap[keyV] = h
	case IEunmType:
		p.keyMap[int32(keyV.Number())] = keyV.String()
		p.handlerMap[keyV.String()] = h
	case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
		k := toStr(keyV)
		p.keyMap[toInt32(keyV)] = k
		p.handlerMap[k] = h
	default:
		if v, ok := key.(fmt.Stringer); ok {
			p.handlerMap[v.String()] = h
		} else {
			panic(fmt.Sprintf("Unavailable types: %T", key))
		}
	}
}

func (p *Processor) GetHandler(key interface{}) (h Handler, ok bool) {
	var k string
	switch keyV := key.(type) {
	case string:
		k = keyV
	case IEunmType:
		k, ok = p.keyMap[int32(keyV.Number())]
		if !ok {
			return nil, false
		}
	case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64:
		k, ok = p.keyMap[toInt32(keyV)]
		if !ok {
			return nil, false
		}
	default:
		if v, ok := key.(fmt.Stringer); ok {
			k = v.String()
		} else {
			panic(fmt.Sprintf("Unavailable types: %T", key))
		}
	}
	h, ok = p.handlerMap[k]
	return
}

func (p *Processor) AddHandle(key interface{}, cb func(val interface{}) error) *Processor {
	p.SetHandler(key, NewAny(cb))
	return p
}

func (p *Processor) AddNone(key interface{}, cb func() error) *Processor {
	p.SetHandler(key, NewNone(cb))
	return p
}

func (p *Processor) AddBool(key interface{}, cb func(val bool) error) *Processor {
	p.SetHandler(key, NewBool(cb))
	return p
}

func (p *Processor) AddString(key interface{}, cb func(val string) error) *Processor {
	p.SetHandler(key, NewString(cb, false))
	return p
}

func (p *Processor) AddStringIgnoreZero(key interface{}, cb func(val string) error) *Processor {
	p.SetHandler(key, NewString(cb, true))
	return p
}

func (p *Processor) AddStringList(key interface{}, cb func(val []string) error) *Processor {
	p.SetHandler(key, NewStringList(cb, false))
	return p
}

func (p *Processor) AddStringListIgnoreZero(key interface{}, cb func(val []string) error) *Processor {
	p.SetHandler(key, NewStringList(cb, true))
	return p
}

func (p *Processor) AddInt32(key interface{}, cb func(val int32) error) *Processor {
	p.SetHandler(key, NewInt32(cb, false))
	return p
}

func (p *Processor) AddInt32IgnoreZero(key interface{}, cb func(val int32) error) *Processor {
	p.SetHandler(key, NewInt32(cb, true))
	return p
}

func (p *Processor) AddInt32List(key interface{}, cb func(valList []int32) error) *Processor {
	p.SetHandler(key, NewInt32List(cb, false))
	return p
}

func (p *Processor) AddInt32ListIgnoreZero(key interface{}, cb func(valList []int32) error) *Processor {
	p.SetHandler(key, NewInt32List(cb, true))
	return p
}

func (p *Processor) AddInt32Range(key interface{}, cb func(begin, end int32) error) *Processor {
	p.SetHandler(key, NewInt32Range(cb, false))
	return p
}

func (p *Processor) AddInt32RangeIgnoreZero(key interface{}, cb func(begin, end int32) error) *Processor {
	p.SetHandler(key, NewInt32Range(cb, true))
	return p
}

func (p *Processor) AddUint32(key interface{}, cb func(val uint32) error) *Processor {
	p.SetHandler(key, NewUint32(cb, false))
	return p
}

func (p *Processor) AddUint32IgnoreZero(key interface{}, cb func(val uint32) error) *Processor {
	p.SetHandler(key, NewUint32(cb, true))
	return p
}

func (p *Processor) AddUint32List(key interface{}, cb func(valList []uint32) error) *Processor {
	p.SetHandler(key, NewUint32List(cb, false))
	return p
}

func (p *Processor) AddUint32ListIgnoreZero(key interface{}, cb func(valList []uint32) error) *Processor {
	p.SetHandler(key, NewUint32List(cb, true))
	return p
}

func (p *Processor) AddUint32Range(key interface{}, cb func(begin, end uint32) error) *Processor {
	p.SetHandler(key, NewUint32Range(cb, false))
	return p
}

func (p *Processor) AddUint32RangeIgnoreZero(key interface{}, cb func(begin, end uint32) error) *Processor {
	p.SetHandler(key, NewUint32Range(cb, true))
	return p
}

func (p *Processor) AddInt64(key interface{}, cb func(val int64) error) *Processor {
	p.SetHandler(key, NewInt64(cb, false))
	return p
}

func (p *Processor) AddInt64IgnoreZero(key interface{}, cb func(val int64) error) *Processor {
	p.SetHandler(key, NewInt64(cb, true))
	return p
}

func (p *Processor) AddInt64List(key interface{}, cb func(valList []int64) error) *Processor {
	p.SetHandler(key, NewInt64List(cb, false))
	return p
}

func (p *Processor) AddInt64ListIgnoreZero(key interface{}, cb func(valList []int64) error) *Processor {
	p.SetHandler(key, NewInt64List(cb, true))
	return p
}

func (p *Processor) AddInt64Range(key interface{}, cb func(begin, end int64) error) *Processor {
	p.SetHandler(key, NewInt64Range(cb, false))
	return p
}

func (p *Processor) AddInt64RangeIgnoreZero(key interface{}, cb func(begin, end int64) error) *Processor {
	p.SetHandler(key, NewInt64Range(cb, false))
	return p
}

func (p *Processor) AddUint64(key interface{}, cb func(val uint64) error) *Processor {
	p.SetHandler(key, NewUint64(cb, false))
	return p
}

func (p *Processor) AddUint64IgnoreZero(key interface{}, cb func(val uint64) error) *Processor {
	p.SetHandler(key, NewUint64(cb, true))
	return p
}

func (p *Processor) AddUint64List(key interface{}, cb func(valList []uint64) error) *Processor {
	p.SetHandler(key, NewUint64List(cb, false))
	return p
}

func (p *Processor) AddUint64ListIgnoreZero(key interface{}, cb func(valList []uint64) error) *Processor {
	p.SetHandler(key, NewUint64List(cb, true))
	return p
}

func (p *Processor) AddUint64Range(key interface{}, cb func(begin, end uint64) error) *Processor {
	p.SetHandler(key, NewUint64Range(cb, false))
	return p
}

func (p *Processor) AddUint64RangeIgnoreZero(key interface{}, cb func(begin, end uint64) error) *Processor {
	p.SetHandler(key, NewUint64Range(cb, true))
	return p
}

func (p *Processor) Process(opts interface{}) error {
	if options, ok := opts.(*Options); ok {
		return p.ProcessOptions(options)
	}
	return p.ProcessStruct(opts)
}

func (p *Processor) ProcessStruct(obj interface{}) error {
	T := reflect.TypeOf(obj)
	if T.Kind() == reflect.Ptr {
		T = T.Elem()
	}
	if T.Kind() != reflect.Struct {
		return fmt.Errorf("requires a structure object, but the object type is %T", obj)
	}
	if len(p.handlerMap) == 0 {
		return nil
	}

	var err error
	optsV := reflect.ValueOf(obj)
	if optsV.Kind() == reflect.Ptr {
		optsV = optsV.Elem()
	}
	for i := 0; i < optsV.NumField(); i++ {
		fieldV := optsV.Field(i)
		if fieldV.Kind() == reflect.Ptr {
			if fieldV.IsNil() {
				continue
			}
			fieldV = fieldV.Elem()
		}
		if p.skipZero && fieldV.IsZero() {
			continue
		}

		fieldT := optsV.Type().Field(i)
		h, ok := p.GetHandler(fieldT.Name)
		if !ok {
			continue
		}
		err = h.Apply(fieldV.Interface())
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Processor) ProcessOptions(options *Options) error {
	if options == nil || len(p.handlerMap) == 0 {
		return nil
	}

	for _, v := range options.Options {
		h, ok := p.GetHandler(v.Kind)
		if !ok {
			continue
		}
		if p.skipZero && reflect.ValueOf(v.Value).IsZero() {
			continue
		}
		err := h.Apply(v.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func toInt32(key interface{}) int32 {
	switch x := key.(type) {
	case int:
		return int32(x)
	case int8:
		return int32(x)
	case int16:
		return int32(x)
	case int32:
		return x
	case int64:
		return int32(x)
	case uint:
		return int32(x)
	case uint8:
		return int32(x)
	case uint16:
		return int32(x)
	case uint32:
		return int32(x)
	case uint64:
		return int32(x)
	case bool:
		if x {
			return 1
		} else {
			return 0
		}
	case string:
		i, _ := strconv.ParseInt(x, 10, 32)
		return int32(i)
	case float32:
		return int32(x)
	case float64:
		return int32(x)
	default:
		return 0
	}
}

func toStr(v interface{}) string {
	if v == nil {
		return ""
	}
	switch vv := v.(type) {
	case int:
		return strconv.FormatInt(int64(vv), 10)
	case int8:
		return strconv.FormatInt(int64(vv), 10)
	case int16:
		return strconv.FormatInt(int64(vv), 10)
	case int32:
		return strconv.FormatInt(int64(vv), 10)
	case int64:
		return strconv.FormatInt(vv, 10)
	case uint:
		return strconv.FormatUint(uint64(vv), 10)
	case uint8:
		return strconv.FormatUint(uint64(vv), 10)
	case uint16:
		return strconv.FormatUint(uint64(vv), 10)
	case uint32:
		return strconv.FormatUint(uint64(vv), 10)
	case uint64:
		return strconv.FormatUint(vv, 10)
	case string:
		return vv
	case bool:
		if vv {
			return "true"
		}
		return "false"

	}
	return fmt.Sprintf("%v", v)
}
