package optx

import (
	"fmt"
	"reflect"
)

type handleFunc = func(val interface{}) error

type PtrProcessor struct {
	ptrHandlerMap map[string]handleFunc
}

func NewPtrProcessor() *PtrProcessor {
	return &PtrProcessor{
		ptrHandlerMap: make(map[string]handleFunc),
	}
}
func (p *PtrProcessor) AddHandle(key interface{}, h handleFunc) *PtrProcessor {
	p.ptrHandlerMap[toString(key)] = h
	return p
}

func toString(key interface{}) string {
	switch x := key.(type) {
	case string:
		return x
	case *string:
		return *x
	default:
		return fmt.Sprintf("%v", x)
	}
}

func (p *PtrProcessor) Process(opts interface{}) error {
	if len(p.ptrHandlerMap) == 0 {
		return nil
	}

	var err error
	optsV := reflect.ValueOf(opts)
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

		fieldT := optsV.Type().Field(i)
		h := p.ptrHandlerMap[fieldT.Name]
		if h == nil {
			continue
		}
		err = h(fieldV.Interface())
		if err != nil {
			return err
		}
	}
	return nil
}
