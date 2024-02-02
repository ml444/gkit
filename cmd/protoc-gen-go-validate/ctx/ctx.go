package ctx

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/funcs"
	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/v"
)

type ValidateCtx struct {
	Imports      []*ImportCtx
	Messages     []*MessageCtx
	NeedWellKnow *NeedWellKnown
	NeedCommon   bool
}

type NeedWellKnown struct {
	Email    bool
	Hostname bool
	UUID     bool
}

type ImportCtx struct {
	Alias string
	Path  string
}

type MessageCtx struct {
	Desc           protoreflect.MessageDescriptor
	TypeName       string
	Fields         []*FieldCtx
	Required       bool
	Disabled       bool
	Ignored        bool
	NonOneOfFields []*FieldCtx
	OptionalFields []*FieldCtx
	RealOneOfs     map[string]*OneOfField
	SubMessageCtxs []*MessageCtx
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

type FieldCtx struct {
	//MessageDesc protoreflect.MessageDescriptor
	Desc     protoreflect.FieldDescriptor
	Field    *protogen.Field
	Rules    proto.Message
	Pkg      string
	Name     string
	Type     string
	Index    int
	OnKey    string
	Required bool
	Skip     bool
	Wrap     string
	TmplName string
	Err      error
	Accessor string
	ErrCode  int32
}

func (fc *FieldCtx) FullName() string {
	isEnum := fc.Desc.Kind() == protoreflect.EnumKind
	if isEnum {
		return funcs.EnumName(fc.Desc.Enum())
	}

	if fc.Pkg != "" {
		return fc.Pkg + "." + fc.Name
	}
	return fc.Name
}

// ImpCtx returns the import package name of the field
func (fc *FieldCtx) ImpCtx() *ImportCtx {
	if len(funcs.ExtraPkg) == 0 {
		funcs.GetImports(funcs.FileDescriptor(fc.Desc.Parent()))
	}
	var parent protoreflect.Descriptor
	switch fc.Desc.Kind() {
	case protoreflect.MessageKind:
		parent = fc.Desc.Message().Parent()
	case protoreflect.EnumKind:
		parent = fc.Desc.Enum().Parent()
	default:
		return nil
	}
	for {
		message, ok := parent.(protoreflect.MessageDescriptor)
		if ok {
			parent = message.Parent()
		} else {
			if pkgName, ok := funcs.ExtraPkg[string(parent.FullName())]; ok {
				fc.Pkg = pkgName
				return &ImportCtx{
					Alias: pkgName,
					Path:  funcs.ExtraPkgPath[pkgName],
				}
			}
			break
		}
	}
	return nil
}

func (fc *FieldCtx) GetTmplName() (string, error) {
	if fc.Err != nil {
		return "", fc.Err
	}
	return fc.TmplName, nil
}

func (fc *FieldCtx) Elem(def string) (elem FieldCtx, err error) {
	elem = *fc
	elem.Accessor = def
	var rules *v.FieldRules
	switch r := fc.Rules.(type) {
	case *v.MapRules:
		rules = r.GetValues()
	case *v.RepeatedRules:
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

func (fc *FieldCtx) Unwrap(name string) (out FieldCtx, err error) {
	if fc.TmplName != "wrapper" {
		err = fmt.Errorf("cannot unwrap non-wrapper type %q", fc.TmplName)
		return
	}

	return FieldCtx{
		Field:    fc.Field,
		Rules:    fc.Rules,
		TmplName: fc.Wrap,
		Accessor: name,
	}, nil
}

func (fc FieldCtx) GetAccessor() string {
	if fc.Accessor != "" {
		if fc.Accessor == "wrapper" {
			return "wrapper.GetValue()"
		}
		return fc.Accessor
	}

	return fmt.Sprintf("m.Get%s()", fc.Field.GoName)
}
func (fc *FieldCtx) SetAccessor(def string) *FieldCtx {
	fc.Accessor = def
	return fc
}
func (fc *FieldCtx) MapTypeName(field *protogen.Field, desc protoreflect.FieldDescriptor, rule *v.FieldRules) *FieldCtx {
	rType, rIns, mRule, wrap := ResolveRules(desc, rule)
	elem := FieldCtx{
		Desc:     desc,
		Field:    field,
		Rules:    rIns,
		TmplName: rType,
	}
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
	return &elem
}
