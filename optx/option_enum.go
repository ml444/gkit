package optx

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type Processor struct {
	handlerMap map[int32]handler
	skipZero   bool
}

func NewEnumProcessor() *Processor {
	return &Processor{
		handlerMap: make(map[int32]handler),
	}
}

func (p *Processor) SkipZero() *Processor {
	p.skipZero = true
	return p
}

func (p *Processor) setHandler(key interface{}, h handler) {
	p.handlerMap[toInt32(key)] = h
}

func (p *Processor) AddNone(key interface{}, cb func() error) *Processor {
	p.setHandler(key, NewNone(cb))
	return p
}

func (p *Processor) AddBool(key interface{}, cb func(val bool) error) *Processor {
	p.setHandler(key, NewBool(cb))
	return p
}

func (p *Processor) AddInt32(key interface{}, cb func(val int32) error) *Processor {
	p.setHandler(key, NewInt32(cb, false))
	return p
}

func (p *Processor) AddInt32IgnoreZero(key interface{}, cb func(val int32) error) *Processor {
	p.setHandler(key, NewInt32(cb, true))
	return p
}

func (p *Processor) AddInt32List(key interface{}, cb func(valList []int32) error) *Processor {
	p.setHandler(key, NewInt32List(cb, false))
	return p
}

func (p *Processor) AddInt32ListIgnoreZero(key interface{}, cb func(valList []int32) error) *Processor {
	p.setHandler(key, NewInt32List(cb, true))
	return p
}

func (p *Processor) AddInt32Range(key interface{}, cb func(begin, end int32) error) *Processor {
	p.setHandler(key, NewInt32Range(cb, false))
	return p
}

func (p *Processor) AddInt32RangeIgnoreZero(key interface{}, cb func(begin, end int32) error) *Processor {
	p.setHandler(key, NewInt32Range(cb, true))
	return p
}

func (p *Processor) AddString(key interface{}, cb func(val string) error) *Processor {
	p.setHandler(key, NewString(cb, false))
	return p
}

func (p *Processor) AddStringIgnoreZero(key interface{}, cb func(val string) error) *Processor {
	p.setHandler(key, NewString(cb, true))
	return p
}

func (p *Processor) AddStringList(key interface{}, cb func(val []string) error) *Processor {
	p.setHandler(key, NewStringList(cb, false))
	return p
}

func (p *Processor) AddStringListIgnoreZero(key interface{}, cb func(val []string) error) *Processor {
	p.setHandler(key, NewStringList(cb, true))
	return p
}

func (p *Processor) AddUint32(key interface{}, cb func(val uint32) error) *Processor {
	p.setHandler(key, NewUint32(cb, false))
	return p
}

func (p *Processor) AddUint32IgnoreZero(key interface{}, cb func(val uint32) error) *Processor {
	p.setHandler(key, NewUint32(cb, true))
	return p
}

func (p *Processor) AddUint32List(key interface{}, cb func(valList []uint32) error) *Processor {
	p.setHandler(key, NewUint32List(cb, false))
	return p
}

func (p *Processor) AddUint32ListIgnoreZero(key interface{}, cb func(valList []uint32) error) *Processor {
	p.setHandler(key, NewUint32List(cb, true))
	return p
}

func (p *Processor) AddUint32Range(key interface{}, cb func(begin, end uint32) error) *Processor {
	p.setHandler(key, NewUint32Range(cb, false))
	return p
}

func (p *Processor) AddUint32RangeIgnoreZero(key interface{}, cb func(begin, end uint32) error) *Processor {
	p.setHandler(key, NewUint32Range(cb, true))
	return p
}

func (p *Processor) AddUint64(key interface{}, cb func(val uint64) error) *Processor {
	p.setHandler(key, NewUint64(cb, false))
	return p
}

func (p *Processor) AddUint64IgnoreZero(key interface{}, cb func(val uint64) error) *Processor {
	p.setHandler(key, NewUint64(cb, true))
	return p
}

func (p *Processor) AddUint64List(key interface{}, cb func(valList []uint64) error) *Processor {
	p.setHandler(key, NewUint64List(cb, false))
	return p
}

func (p *Processor) AddUint64ListIgnoreZero(key interface{}, cb func(valList []uint64) error) *Processor {
	p.setHandler(key, NewUint64List(cb, true))
	return p
}

func (p *Processor) AddUint64Range(key interface{}, cb func(begin, end uint64) error) *Processor {
	p.setHandler(key, NewUint64Range(cb, false))
	return p
}

func (p *Processor) AddUint64RangeIgnoreZero(key interface{}, cb func(begin, end uint64) error) *Processor {
	p.setHandler(key, NewUint64Range(cb, true))
	return p
}

func (p *Processor) AddInt64(key interface{}, cb func(val int64) error) *Processor {
	p.setHandler(key, NewInt64(cb, false))
	return p
}

func (p *Processor) AddInt64IgnoreZero(key interface{}, cb func(val int64) error) *Processor {
	p.setHandler(key, NewInt64(cb, true))
	return p
}

func (p *Processor) AddInt64List(key interface{}, cb func(valList []int64) error) *Processor {
	p.setHandler(key, NewInt64List(cb, false))
	return p
}

func (p *Processor) AddInt64ListIgnoreZero(key interface{}, cb func(valList []int64) error) *Processor {
	p.setHandler(key, NewInt64List(cb, true))
	return p
}

func (p *Processor) AddInt64Range(key interface{}, cb func(begin, end int64) error) *Processor {
	p.setHandler(key, NewInt64Range(cb, false))
	return p
}

func (p *Processor) AddInt64RangeIgnoreZero(key interface{}, cb func(begin, end int64) error) *Processor {
	p.setHandler(key, NewInt64Range(cb, false))
	return p
}

func (p *Processor) Process(options *Options) error {
	if options == nil || len(p.handlerMap) == 0 {
		return nil
	}

	for _, v := range options.Options {
		h := p.handlerMap[v.Kind]
		if h == nil {
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

func inSliceStr(s string, list []string) bool {
	for _, v := range list {
		if s == v {
			return true
		}
	}
	return false
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
	default:
		return 0
	}
}

func toStr(i interface{}) string {
	t := reflect.TypeOf(i)
	k := t.Kind()
	switch k {
	case reflect.Int,
		reflect.Int32,
		reflect.Int64,
		reflect.Int16,
		reflect.Int8:
		return strconv.FormatInt(reflect.ValueOf(i).Int(), 10)
	case reflect.Uint,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uint16,
		reflect.Uint8:
		return strconv.FormatUint(reflect.ValueOf(i).Uint(), 10)
	case reflect.String:
		return reflect.ValueOf(i).String()
	case reflect.Bool:
		if reflect.ValueOf(i).Bool() {
			return "1"
		} else {
			return "0"
		}
	}
	return fmt.Sprintf("%v", i)
}

func NewEnumOptions(opts ...interface{}) *Options {
	if len(opts)%2 != 0 {
		panic(fmt.Sprintf("invalid number of opts argument %d", len(opts)))
	}
	l := &Options{}
	for i := 0; i < len(opts); i += 2 {
		l.AddOpt(opts[i], opts[i+1])
	}
	return l
}

func (opt *Options) IsOptExist(enumkey interface{}) bool {
	int32Key := toInt32(enumkey)

	for _, opt := range opt.Options {
		if opt.Kind == int32Key {
			return true
		}
	}

	return false
}

func (opt *Options) GetOptValue(enumkey interface{}) (string, bool) {
	int32Key := toInt32(enumkey)
	for _, opt := range opt.Options {
		if opt.Kind == int32Key {
			return opt.Value, true
		}
	}

	return "", false
}

func (opt *Options) AddOpt(enumkey, val interface{}) *Options {
	int32Key := toInt32(enumkey)

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
	opt.Options = append(opt.Options,
		&Options_Option{Kind: int32Key, Value: strVal})
	return opt
}
