package listoption

import (
	"fmt"
	"github.com/ml444/gkit/errors"
	"github.com/ml444/gkit/logger"
	"log"
	"reflect"
	"strconv"
	"strings"
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
	typ             int
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
type Processor struct {
	listOption *ListOption
	handlers   map[int32]*Handler
	logger     logger.Logger
}

func toInt32(i interface{}) int32 {
	t := reflect.TypeOf(i)
	k := t.Kind()
	switch k {
	case reflect.Int,
		reflect.Int32,
		reflect.Int64,
		reflect.Int16,
		reflect.Int8:
		return int32(reflect.ValueOf(i).Int())
	case reflect.Uint,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uint16,
		reflect.Uint8:
		return int32(reflect.ValueOf(i).Uint())
	}
	return 0
}

func NewProcessor(listOption *ListOption) *Processor {
	return &Processor{
		listOption: listOption,
		handlers:   make(map[int32]*Handler),
	}
}
func (p *Processor) SetLogger(l logger.Logger) {
	p.logger = l
}

func (p *Processor) AddNone(typ interface{}, cb func() error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:    ValueTypeNil,
		cbNone: cb,
	}
	return p
}
func (p *Processor) AddBool(typ interface{}, cb func(val bool) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:    ValueTypeBool,
		cbBool: cb,
	}
	return p
}

func (p *Processor) AddInt32(typ interface{}, cb func(val int32) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:     ValueTypeInt32,
		cbInt32: cb,
	}
	return p
}
func (p *Processor) AddInt32List(typ interface{}, cb func(valList []int32) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:             ValueTypeUint32List,
		cbInt32List:     cb,
		ignoreZeroValue: true,
	}
	return p
}
func (p *Processor) AddInt32Range(typ interface{}, cb func(begin, end int32) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:             ValueTypeUint32Range,
		cbInt32Range:    cb,
		ignoreZeroValue: true,
	}
	return p
}

func (p *Processor) AddString(typ interface{}, cb func(val string) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:      ValueTypeString,
		cbString: cb,
	}
	return p
}
func (p *Processor) AddStringIgnoreZero(typ interface{}, cb func(val string) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:             ValueTypeString,
		cbString:        cb,
		ignoreZeroValue: true,
	}
	return p
}
func (p *Processor) AddStringList(typ interface{}, cb func(val []string) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:             ValueTypeStringList,
		cbStringList:    cb,
		ignoreZeroValue: true,
	}
	return p
}

func (p *Processor) AddUint32(typ interface{}, cb func(val uint32) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:      ValueTypeUint32,
		cbUint32: cb,
	}
	return p
}
func (p *Processor) AddUint32List(typ interface{}, cb func(valList []uint32) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:             ValueTypeUint32List,
		cbUint32List:    cb,
		ignoreZeroValue: true,
	}
	return p
}
func (p *Processor) AddUint32Range(typ interface{}, cb func(begin, end uint32) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:             ValueTypeUint32Range,
		cbUint32Range:   cb,
		ignoreZeroValue: true,
	}
	return p
}

func (p *Processor) AddUint64(typ interface{}, cb func(val uint64) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:      ValueTypeUint64,
		cbUint64: cb,
	}
	return p
}
func (p *Processor) AddUint64List(typ interface{}, cb func(valList []uint64) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:             ValueTypeUint64List,
		cbUint64List:    cb,
		ignoreZeroValue: true,
	}
	return p
}
func (p *Processor) AddUint64Range(typ interface{}, cb func(begin, end uint64) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:             ValueTypeUint32Range,
		cbUint64Range:   cb,
		ignoreZeroValue: true,
	}
	return p
}

func (p *Processor) AddInt64(typ interface{}, cb func(val int64) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:     ValueTypeUint64,
		cbInt64: cb,
	}
	return p
}
func (p *Processor) AddInt64List(typ interface{}, cb func(valList []int64) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:             ValueTypeUint64List,
		cbInt64List:     cb,
		ignoreZeroValue: true,
	}
	return p
}
func (p *Processor) AddInt64Range(typ interface{}, cb func(begin, end int64) error) *Processor {
	x := toInt32(typ)
	p.handlers[x] = &Handler{
		typ:             ValueTypeUint32Range,
		cbInt64Range:    cb,
		ignoreZeroValue: true,
	}
	return p
}

func (p *Processor) Process() error {
	if p.listOption == nil || p.handlers == nil || len(p.handlers) == 0 {
		return nil
	}
	newError := func(typ int32, expectType string) error {
		return errors.CreateErrorf(errors.DefaultStatusCode, errors.ErrCodeInvalidParamSys,
			fmt.Sprintf("invalid option value with type %d, expected %s", typ, expectType))
	}
	var err error
	for _, v := range p.listOption.Options {
		h := p.handlers[v.Type]
		if h == nil {
			continue
		}
		switch h.typ {
		case ValueTypeNil:
			if h.cbNone != nil {
				err = h.cbNone()
				if err != nil {
					return err
				}
			}

		case ValueTypeBool:
			//if v.Value != "0" && v.Value != "1" {
			//	continue
			//}
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
				return newError(v.Type, "int32")
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
					return newError(v.Type, "int32")
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
				if p.logger != nil {
					p.logger.Warnf("the value:[] does not meet the requirements. skip it.", v.Value)
				}
				continue
			}
			t1, err := strconv.ParseInt(tStr[0], 10, 64)
			if err != nil {
				return newError(v.Type, "int32")
			}
			t2, err := strconv.ParseInt(tStr[1], 10, 64)
			if err != nil {
				return newError(v.Type, "int32")
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
				return newError(v.Type, "uint32")
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
					return newError(v.Type, "uint32")
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
				if p.logger != nil {
					p.logger.Warnf("the value:[] does not meet the requirements. skip it.", v.Value)
				}
				continue
			}
			t1, err := strconv.ParseUint(tStr[0], 10, 64)
			if err != nil {
				return newError(v.Type, "uint32")
			}
			t2, err := strconv.ParseUint(tStr[1], 10, 64)
			if err != nil {
				return newError(v.Type, "uint32")
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
				return newError(v.Type, "uint64")
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
					return newError(v.Type, "uint64")
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
				if p.logger != nil {
					p.logger.Warnf("the value:[] does not meet the requirements. skip it.", v.Value)
				}
				continue
			}
			t1, err := strconv.ParseUint(tStr[0], 10, 64)
			if err != nil {
				return newError(v.Type, "uint64")
			}
			t2, err := strconv.ParseUint(tStr[1], 10, 64)
			if err != nil {
				return newError(v.Type, "uint64")
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
				return newError(v.Type, "int64")
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
					return newError(v.Type, "int64")
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
				if p.logger != nil {
					p.logger.Warnf("the value:[] does not meet the requirements. skip it.", v.Value)
				}
				continue
			}
			t1, err := strconv.ParseInt(tStr[0], 10, 64)
			if err != nil {
				return newError(v.Type, "int64")
			}
			t2, err := strconv.ParseInt(tStr[1], 10, 64)
			if err != nil {
				return newError(v.Type, "int64")
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

func NewListOption(opts ...interface{}) *ListOption {
	if len(opts)%2 != 0 {
		log.Panicf("invalid number of opts argument %d", len(opts))
	}
	l := &ListOption{}
	for i := 0; i < len(opts); i += 2 {
		l.AddOpt(opts[i], opts[i+1])
	}
	return l
}

func (opt *ListOption) SetLimit(limit uint32) *ListOption {
	opt.Limit = limit
	return opt
}

func (opt *ListOption) SetOffset(offset uint32) *ListOption {
	opt.Offset = offset
	return opt
}

func (opt *ListOption) IsOptExist(typ interface{}) bool {
	var typInt int32
	if reflect.TypeOf(typ).Kind() == reflect.Uint32 {
		typInt = int32(reflect.ValueOf(typ).Uint())
	} else {
		typInt = int32(reflect.ValueOf(typ).Int())
	}
	if typInt <= 0 {
		log.Panicf("invalid type %d", typ)
	}

	for _, opt := range opt.Options {
		if opt.Type == typInt {
			return true
		}
	}

	return false
}

func (opt *ListOption) GetOptValue(typ interface{}) (string, bool) {
	var typInt int32
	if reflect.TypeOf(typ).Kind() == reflect.Uint32 {
		typInt = int32(reflect.ValueOf(typ).Uint())
	} else {
		typInt = int32(reflect.ValueOf(typ).Int())
	}
	if typInt <= 0 {
		log.Panicf("invalid type %d", typ)
	}

	for _, opt := range opt.Options {
		if opt.Type == typInt {
			return opt.Value, true
		}
	}

	return "", false
}

func (opt *ListOption) AddOptIf(flag bool, typ, val interface{}) *ListOption {
	if flag {
		opt.AddOpt(typ, val)
	}

	return opt
}

func (opt *ListOption) AddOpt(typ, val interface{}) *ListOption {
	var typInt int32
	if reflect.TypeOf(typ).Kind() == reflect.Uint32 {
		typInt = int32(reflect.ValueOf(typ).Uint())
	} else {
		typInt = int32(reflect.ValueOf(typ).Int())
	}
	if typInt <= 0 {
		log.Panicf("invalid type %d", typ)
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
		&ListOption_Option{Type: typInt, Value: strVal})
	return opt
}

func (opt *ListOption) SetSkipCount() *ListOption {
	opt.SkipCount = true
	return opt
}

func (opt *ListOption) CloneSkipOpts() *ListOption {
	l := NewListOption().
		SetOffset(opt.GetOffset()).
		SetLimit(opt.GetLimit())
	if l.SkipCount {
		l.SetSkipCount()
	}
	return l
}

func getOptTypeFromInterface(typ interface{}) uint32 {
	t := reflect.TypeOf(typ)
	v := reflect.ValueOf(typ)
	if t.Kind() == reflect.Int32 {
		return uint32(v.Int())
	} else if t.Kind() == reflect.Uint32 {
		return uint32(v.Uint())
	} else {
		log.Panicf("unsupported type %s of opt with value %v", t.String(), typ)
	}
	return 0
}

func (opt *ListOption) CloneChangeOptTypes(optPairs ...interface{}) *ListOption {
	l := opt.CloneSkipOpts()
	if len(optPairs)%2 != 0 {
		log.Panicf("invalid number of opts argument %d", len(optPairs))
	}
	kv := map[uint32]uint32{}
	for i := 0; i < len(optPairs); i += 2 {
		typ := optPairs[i]
		val := optPairs[i+1]
		kv[getOptTypeFromInterface(typ)] = getOptTypeFromInterface(val)
	}
	for _, v := range opt.Options {
		t := uint32(v.Type)
		if vv, ok := kv[t]; ok {
			l.AddOpt(vv, v.Value)
		}
	}
	return l
}
