package listoption

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/log"
)

const (
	ValueTypeNil = iota
	ValueTypeBool
	ValueTypeString
	ValueTypeStringList
	ValueTypeInt32
	ValueTypeInt32List
	ValueTypeInt32Range
	ValueTypeUint32
	ValueTypeUint32List
	ValueTypeUint32Range
	ValueTypeInt64
	ValueTypeInt64List
	ValueTypeInt64Range
	ValueTypeUint64
	ValueTypeUint64List
	ValueTypeUint64Range
)

type Handler struct {
	key             int
	cbNone          func() error
	cbInt32         func(val int32) error
	cbInt32List     func(valList []int32) error
	cbInt32Range    func(begin, end int32) error
	cbUint32        func(val uint32) error
	cbUint32List    func(valList []uint32) error
	cbUint32Range   func(begin, end uint32) error
	cbInt64         func(val int64) error
	cbInt64List     func(val []int64) error
	cbInt64Range    func(begin, end int64) error
	cbUint64        func(val uint64) error
	cbUint64List    func(val []uint64) error
	cbUint64Range   func(begin, end uint64) error
	cbBool          func(val bool) error
	cbString        func(val string) error
	cbStringList    func(val []string) error
	ignoreZeroValue bool
}

func (h *Handler) setValue(v interface{}) {}
func (h *Handler) apply() error           { return nil }

type Processor struct {
	listOption *Options
	handlers   map[int32]handler
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

func NewProcessor(listOption *Options) *Processor {
	return &Processor{
		listOption: listOption,
		handlers:   make(map[int32]handler),
	}
}

func (p *Processor) setHandler(key interface{}, h handler) {
	p.handlers[toInt32(key)] = h
}

func (p *Processor) AddNone(key interface{}, cb func() error) *Processor {
	p.setHandler(key, &Handler{
		key:    ValueTypeNil,
		cbNone: cb,
	})
	return p
}
func (p *Processor) AddBool(key interface{}, cb func(val bool) error) *Processor {
	p.setHandler(key, &Handler{
		key:    ValueTypeBool,
		cbBool: cb,
	})
	return p
}

func (p *Processor) AddInt32(key interface{}, cb func(val int32) error) *Processor {
	p.setHandler(key, &Handler{
		key:     ValueTypeInt32,
		cbInt32: cb,
	})
	return p
}
func (p *Processor) AddInt32List(key interface{}, cb func(valList []int32) error) *Processor {
	p.setHandler(key, &Handler{
		key:             ValueTypeUint32List,
		cbInt32List:     cb,
		ignoreZeroValue: true,
	})
	return p
}
func (p *Processor) AddInt32Range(key interface{}, cb func(begin, end int32) error) *Processor {
	p.setHandler(key, &Handler{
		key:             ValueTypeUint32Range,
		cbInt32Range:    cb,
		ignoreZeroValue: true,
	})
	return p
}

func (p *Processor) AddString(key interface{}, cb func(val string) error) *Processor {
	p.setHandler(key, &Handler{
		key:      ValueTypeString,
		cbString: cb,
	})
	return p
}
func (p *Processor) AddStringIgnoreZero(key interface{}, cb func(val string) error) *Processor {
	p.setHandler(key, &Handler{
		key:             ValueTypeString,
		cbString:        cb,
		ignoreZeroValue: true,
	})
	return p
}
func (p *Processor) AddStringList(key interface{}, cb func(val []string) error) *Processor {
	p.setHandler(key, &Handler{
		key:             ValueTypeStringList,
		cbStringList:    cb,
		ignoreZeroValue: true,
	})
	return p
}

func (p *Processor) AddUint32(key interface{}, cb func(val uint32) error) *Processor {
	p.setHandler(key, &Handler{
		key:      ValueTypeUint32,
		cbUint32: cb,
	})
	return p
}
func (p *Processor) AddUint32List(key interface{}, cb func(valList []uint32) error) *Processor {
	p.setHandler(key, &Handler{
		key:             ValueTypeUint32List,
		cbUint32List:    cb,
		ignoreZeroValue: true,
	})
	return p
}
func (p *Processor) AddUint32Range(key interface{}, cb func(begin, end uint32) error) *Processor {
	p.setHandler(key, &Handler{
		key:             ValueTypeUint32Range,
		cbUint32Range:   cb,
		ignoreZeroValue: true,
	})
	return p
}

func (p *Processor) AddUint64(key interface{}, cb func(val uint64) error) *Processor {
	p.setHandler(key, &Handler{
		key:      ValueTypeUint64,
		cbUint64: cb,
	})
	return p
}
func (p *Processor) AddUint64List(key interface{}, cb func(valList []uint64) error) *Processor {
	p.setHandler(key, &Handler{
		key:             ValueTypeUint64List,
		cbUint64List:    cb,
		ignoreZeroValue: true,
	})
	return p
}
func (p *Processor) AddUint64Range(key interface{}, cb func(begin, end uint64) error) *Processor {
	p.setHandler(key, &Handler{
		key:             ValueTypeUint32Range,
		cbUint64Range:   cb,
		ignoreZeroValue: true,
	})
	return p
}

func (p *Processor) AddInt64(key interface{}, cb func(val int64) error) *Processor {
	p.setHandler(key, &Handler{
		key:     ValueTypeUint64,
		cbInt64: cb,
	})
	return p
}
func (p *Processor) AddInt64List(key interface{}, cb func(valList []int64) error) *Processor {
	p.setHandler(key, &Handler{
		key:             ValueTypeUint64List,
		cbInt64List:     cb,
		ignoreZeroValue: true,
	})
	return p
}
func (p *Processor) AddInt64Range(key interface{}, cb func(begin, end int64) error) *Processor {
	p.setHandler(key, &Handler{
		key:             ValueTypeUint32Range,
		cbInt64Range:    cb,
		ignoreZeroValue: true,
	})
	return p
}

func (p *Processor) Process() error {
	if p.listOption == nil || len(p.handlers) == 0 {
		return nil
	}
	newError := func(key int32, expectType string) error {
		return errorx.CreateErrorf(errorx.DefaultStatusCode, errorx.ErrCodeInvalidParamSys,
			fmt.Sprintf("invalid option value with type %d, expected %s", key, expectType))
	}
	var err error
	for _, v := range p.listOption.Options {
		h := p.handlers[v.Kind].(*Handler)
		if h == nil {
			continue
		}
		switch h.key {
		case ValueTypeNil:
			if h.cbNone != nil {
				err = h.cbNone()
				if err != nil {
					return err
				}
			}

		case ValueTypeBool:
			value := strings.ToLower(v.Value)
			var x bool
			if inSliceStr(value, []string{"1", "true"}) {
				x = true
			} else if inSliceStr(value, []string{"0", "false"}) {
				x = false
			} else {
				continue
			}
			if h.cbBool != nil {
				err = h.cbBool(x)
				if err != nil {
					return err
				}
			}

		case ValueTypeString:
			if v.Value == "" && h.ignoreZeroValue {
				continue
			}
			if h.cbString != nil {
				if err = h.cbString(v.Value); err != nil {
					return err
				}
			}

		case ValueTypeStringList:
			if v.Value == "" && h.ignoreZeroValue {
				continue
			}
			if h.cbStringList != nil {
				list := strings.Split(v.Value, ",")
				// 过滤掉空串
				var nonEmptyList []string
				for _, v := range list {
					if v != "" {
						nonEmptyList = append(nonEmptyList, v)
					}
				}
				if len(nonEmptyList) == 0 && h.ignoreZeroValue {
					continue
				}
				if err = h.cbStringList(nonEmptyList); err != nil {
					return err
				}
			}

		case ValueTypeInt32:
			if v.Value == "" {
				continue
			}
			x, err := strconv.ParseInt(v.Value, 10, 32)
			if err != nil {
				return newError(v.Kind, "int32")
			}
			if h.cbInt32 != nil {
				err = h.cbInt32(int32(x))
				if err != nil {
					return err
				}
			}

		case ValueTypeInt32List:
			if v.Value == "" && h.ignoreZeroValue {
				continue
			}
			list := strings.Split(v.Value, ",")
			var intList []int32
			for _, item := range list {
				x, err := strconv.ParseInt(item, 10, 32)
				if err != nil {
					return newError(v.Kind, "int32")
				}
				intList = append(intList, int32(x))
			}
			if h.cbInt32List != nil {
				if err = h.cbInt32List(intList); err != nil {
					return err
				}
			}

		case ValueTypeInt32Range:
			tStr, isContinue := toSplitStr(v.Value, h.ignoreZeroValue)
			if isContinue {
				log.Warnf("the value:[] does not meet the requirements. skip it.", v.Value)
				continue
			}
			t1, err := strconv.ParseInt(tStr[0], 10, 64)
			if err != nil {
				return newError(v.Kind, "int32")
			}
			t2, err := strconv.ParseInt(tStr[1], 10, 64)
			if err != nil {
				return newError(v.Kind, "int32")
			}
			if h.cbInt32Range != nil {
				if err = h.cbInt32Range(int32(t1), int32(t2)); err != nil {
					return err
				}
			}

		case ValueTypeUint32:
			if v.Value == "" {
				continue
			}
			x, err := strconv.ParseUint(v.Value, 10, 32)
			if err != nil {
				return newError(v.Kind, "uint32")
			}
			if h.cbUint32 != nil {
				err = h.cbUint32(uint32(x))
				if err != nil {
					return err
				}
			}

		case ValueTypeUint32List:
			if v.Value == "" && h.ignoreZeroValue {
				continue
			}
			list := strings.Split(v.Value, ",")
			var intList []uint32
			for _, item := range list {
				x, err := strconv.ParseUint(item, 10, 32)
				if err != nil {
					return newError(v.Kind, "uint32")
				}
				intList = append(intList, uint32(x))
			}
			if h.cbUint32List != nil {
				if err = h.cbUint32List(intList); err != nil {
					return err
				}
			}

		case ValueTypeUint32Range:
			tStr, isContinue := toSplitStr(v.Value, h.ignoreZeroValue)
			if isContinue {
				log.Warnf("the value:[] does not meet the requirements. skip it.", v.Value)
				continue
			}
			t1, err := strconv.ParseUint(tStr[0], 10, 64)
			if err != nil {
				return newError(v.Kind, "uint32")
			}
			t2, err := strconv.ParseUint(tStr[1], 10, 64)
			if err != nil {
				return newError(v.Kind, "uint32")
			}
			if h.cbUint32Range != nil {
				if err = h.cbUint32Range(uint32(t1), uint32(t2)); err != nil {
					return err
				}
			}

		case ValueTypeUint64:
			if v.Value == "" {
				continue
			}
			x, err := strconv.ParseUint(v.Value, 10, 64)
			if err != nil {
				return newError(v.Kind, "uint64")
			}
			if h.cbUint64 != nil {
				err = h.cbUint64(x)
				if err != nil {
					return err
				}
			}

		case ValueTypeUint64List:
			if v.Value == "" && h.ignoreZeroValue {
				continue
			}
			list := strings.Split(v.Value, ",")
			var intList []uint64
			for _, item := range list {
				x, err := strconv.ParseUint(item, 10, 64)
				if err != nil {
					return newError(v.Kind, "uint64")
				}
				intList = append(intList, x)
			}
			if h.cbUint64List != nil {
				if err = h.cbUint64List(intList); err != nil {
					return err
				}
			}

		case ValueTypeUint64Range:
			tStr, isContinue := toSplitStr(v.Value, h.ignoreZeroValue)
			if isContinue {
				log.Warnf("the value:[] does not meet the requirements. skip it.", v.Value)
				continue
			}
			t1, err := strconv.ParseUint(tStr[0], 10, 64)
			if err != nil {
				return newError(v.Kind, "uint64")
			}
			t2, err := strconv.ParseUint(tStr[1], 10, 64)
			if err != nil {
				return newError(v.Kind, "uint64")
			}
			if h.cbUint64Range != nil {
				if err = h.cbUint64Range(t1, t2); err != nil {
					return err
				}
			}

		case ValueTypeInt64:
			if v.Value == "" {
				continue
			}
			x, err := strconv.ParseInt(v.Value, 10, 32)
			if err != nil {
				return newError(v.Kind, "int64")
			}
			if h.cbInt64 != nil {
				err = h.cbInt64(x)
				if err != nil {
					return err
				}
			}

		case ValueTypeInt64List:
			if v.Value == "" && h.ignoreZeroValue {
				continue
			}
			list := strings.Split(v.Value, ",")
			var intList []int64
			for _, item := range list {
				x, err := strconv.ParseInt(item, 10, 32)
				if err != nil {
					return newError(v.Kind, "int64")
				}
				intList = append(intList, x)
			}
			if h.cbInt64List != nil {
				if err = h.cbInt64List(intList); err != nil {
					return err
				}
			}

		case ValueTypeInt64Range:
			tStr, isContinue := toSplitStr(v.Value, h.ignoreZeroValue)
			if isContinue {
				log.Warnf("the value:[] does not meet the requirements. skip it.", v.Value)
				continue
			}
			t1, err := strconv.ParseInt(tStr[0], 10, 64)
			if err != nil {
				return newError(v.Kind, "int64")
			}
			t2, err := strconv.ParseInt(tStr[1], 10, 64)
			if err != nil {
				return newError(v.Kind, "int64")
			}
			if h.cbInt64Range != nil {
				if err = h.cbInt64Range(t1, t2); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func toSplitStr(s string, ignoreZeroValue bool) (tStr []string, isContinue bool) {
	if strings.Index(s, ",") > 0 {
		tStr = strings.Split(s, ",")
		if len(tStr) != 2 && ignoreZeroValue {
			return tStr, true
		}
	} else {
		tStr = strings.Split(s, "-")
		if len(tStr) != 2 && ignoreZeroValue {
			return tStr, true
		}
	}
	return tStr, false
}
func inSliceStr(s string, list []string) bool {
	for _, v := range list {
		if s == v {
			return true
		}
	}
	return false
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

func NewListOption(opts ...interface{}) *Options {
	if len(opts)%2 != 0 {
		panic(fmt.Sprintf("invalid number of opts argument %d", len(opts)))
	}
	l := &Options{}
	for i := 0; i < len(opts); i += 2 {
		l.AddOpt(opts[i], opts[i+1])
	}
	return l
}

func (opt *Options) IsOptExist(key interface{}) bool {
	var intKind int32
	if reflect.TypeOf(key).Kind() == reflect.Uint32 {
		intKind = int32(reflect.ValueOf(key).Uint())
	} else {
		intKind = int32(reflect.ValueOf(key).Int())
	}
	if intKind <= 0 {
		panic(fmt.Sprintf("invalid type %d", key))
	}

	for _, opt := range opt.Options {
		if opt.Kind == intKind {
			return true
		}
	}

	return false
}

func (opt *Options) GetOptValue(key interface{}) (string, bool) {
	var intKind int32
	if reflect.TypeOf(key).Kind() == reflect.Uint32 {
		intKind = int32(reflect.ValueOf(key).Uint())
	} else {
		intKind = int32(reflect.ValueOf(key).Int())
	}
	if intKind <= 0 {
		panic(fmt.Sprintf("invalid type %d", key))
	}

	for _, opt := range opt.Options {
		if opt.Kind == intKind {
			return opt.Value, true
		}
	}

	return "", false
}

func (opt *Options) AddOptIf(flag bool, key, val interface{}) *Options {
	if flag {
		opt.AddOpt(key, val)
	}

	return opt
}

func (opt *Options) AddOpt(key, val interface{}) *Options {
	var typInt int32
	if reflect.TypeOf(key).Kind() == reflect.Uint32 {
		typInt = int32(reflect.ValueOf(key).Uint())
	} else {
		typInt = int32(reflect.ValueOf(key).Int())
	}
	if typInt <= 0 {
		panic(fmt.Sprintf("invalid type %d", key))
	}
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
		&Options_Option{Kind: typInt, Value: strVal})
	return opt
}
