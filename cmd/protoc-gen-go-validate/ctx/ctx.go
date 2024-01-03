package ctx

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/validate"
)

type MessageCtx struct {
	Desc           protoreflect.MessageDescriptor
	TypeName       string
	Fields         []*FieldCtx
	Required       bool
	Disabled       bool
	Ignored        bool
	NonOneOfFields []*FieldCtx
	RealOneOfs     map[string]*OneOfField
	OptionalFields []*FieldCtx
}

type FieldCtx struct {
	MessageDesc protoreflect.MessageDescriptor
	Desc        protoreflect.FieldDescriptor
	Field       *protogen.Field
	Rules       proto.Message
	Name        string
	Type        string
	Index       int
	OnKey       string
	Required    bool
	Skip        bool
	Wrap        string
	TmplName    string
	Err         error
	Accessor    string
}

type OneOfField struct {
	Field  *protogen.Field
	Fields []*FieldCtx
	Name   string
	Type   string
	//Index    int
	Required bool
	Skip     bool
	//TmplName string
	//Err      error
}

func (fc *FieldCtx) GetTmplName() (string, error) {
	if fc.Err != nil {
		return "", fc.Err
	}
	return fc.TmplName, nil
}

func (fc FieldCtx) Elem(def string) (elem FieldCtx, err error) {
	elem = fc
	elem.Accessor = def
	var rules *validate.FieldRules
	switch r := fc.Rules.(type) {
	case *validate.MapRules:
		rules = r.GetValues()
	case *validate.RepeatedRules:
		rules = r.GetItems()
	default:
		err = fmt.Errorf("cannot get Elem from %s", fc.Field.GoName)
		return
	}

	rType, rIns, mRule, wrap := ResolveRules(fc.Desc, rules)
	elem.Rules = rIns
	if wrap {
		elem.Wrap = rType
		elem.TmplName = "wrapper"
	} else {
		elem.TmplName = rType
	}

	if mRule != nil {
		if mRule.Required != nil {
			elem.Required = *mRule.Required
		}
		if mRule.Skip != nil {
			elem.Skip = *mRule.Skip
		}
	}
	return
}

func (fc FieldCtx) Unwrap(name string) (out FieldCtx, err error) {
	if fc.TmplName != "wrapper" {
		err = fmt.Errorf("cannot unwrap non-wrapper type %q", fc.TmplName)
		return
	}

	return FieldCtx{
		Field:       fc.Field,
		Rules:       fc.Rules,
		MessageDesc: fc.MessageDesc,
		TmplName:    fc.Wrap,
		Accessor:    name,
	}, nil
}
