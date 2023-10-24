package listoption

import (
	"fmt"
	"reflect"
)

type PtrProcessor struct {
	isPtrOptions  bool
	ptrOptions    interface{}
	ptrHandlerMap map[string]handler
}

func NewPtrProcessor(opts interface{}) *PtrProcessor {
	return &PtrProcessor{
		isPtrOptions:  true,
		ptrOptions:    opts,
		ptrHandlerMap: make(map[string]handler),
	}
}
func (p *PtrProcessor) Add(key interface{}, h handler) *PtrProcessor {
	p.ptrHandlerMap[toString(key)] = h
	return p
}
func (p *PtrProcessor) ProcessPtrOpts() error {
	if !p.isPtrOptions || len(p.ptrHandlerMap) == 0 {
		return nil
	}

	var err error
	optsV := reflect.ValueOf(p.ptrOptions)
	if optsV.Kind() == reflect.Ptr {
		optsV = optsV.Elem()
	}
	for i := 0; i < optsV.NumField(); i++ {
		fieldV := optsV.Field(i)
		if fieldV.Kind() == reflect.Ptr {
			fieldV = fieldV.Elem()
		}
		fieldT := optsV.Type().Field(i)
		h := p.ptrHandlerMap[fieldT.Name]
		if h == nil {
			continue
		}
		h.setValue(fieldV.Interface())
		err = h.apply()
		if err != nil {
			return err
		}
	}
	return nil
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
