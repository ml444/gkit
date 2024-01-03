package funcs

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/ctx"
	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/validate"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//	func GetTmplNameByRules(f protogen.Field) (tmplName string, err error) {
//		out.Field = f
//
//		var rules validate.FieldRules
//		fr, ok := proto.GetExtension(f.Desc.Options(), validate.E_Rules).(*validate.FieldRules)
//
//		var wrapped bool
//		if out.Typ, out.Rules, out.MessageRules, wrapped = resolveRules(f.Type(), &rules); wrapped {
//			out.WrapperTyp = out.Typ
//			out.Typ = "wrapper"
//		}
//
//		if out.Typ == "error" {
//			err = fmt.Errorf("unknown rule type (%T)", rules.Type)
//		}
//
//		return
//	}
func Render(tpl *template.Template) func(tmplName string, data interface{}) (string, error) {
	return func(tmplName string, data interface{}) (string, error) {
		var b bytes.Buffer
		err := tpl.ExecuteTemplate(&b, tmplName, data)
		return b.String(), err
	}
}
func accessor(ctx ctx.FieldCtx) string {
	if ctx.Accessor != "" {
		if ctx.Accessor == "wrapper" {
			return "wrapper.GetValue()"
		}
		return ctx.Accessor
	}

	return fmt.Sprintf("m.Get%s()", ctx.Field.GoName)
}

//func errName(m protoreflect.MessageDescriptor) string {
//	return string(m.Name()) + "ValidationError"
//}

//	func multiErrName(m protogen.Message) string {
//		return m.GoIdent.GoName + "MultiError"
//	}
func lookup(f protogen.Field, name string) string {
	return fmt.Sprintf(
		"_%s_%s_%s",
		f.Desc.Parent().Name(),
		f.GoName,
		name,
	)
}
func errIdxCause(field protogen.Field, idx, cause string, reason ...interface{}) string {
	n := field.GoName
	var fld string
	if idx != "" {
		fld = fmt.Sprintf(`fmt.Sprintf("%s[%%v]", %s)`, n, idx)
		//} else if field.Desc.Index() != 0 {
		//	fld = fmt.Sprintf(`fmt.Sprintf("%s[%%v]", %d)`, n, field.Desc.Index()) // TODO
	} else {
		fld = fmt.Sprintf("%q", n)
	}

	causeFld := ""
	if cause != "nil" && cause != "" {
		causeFld = fmt.Sprintf("cause: %s,", cause)
	}

	keyFld := ""
	if field.Desc.IsMap() {
		keyFld = "key: true,"
	}

	return fmt.Sprintf(`%s{
		field: %s,
		reason: %q,
		%s%s
	}`,
		"ValidationError",
		fld,
		fmt.Sprint(reason...),
		causeFld,
		keyFld)
}

func err(field protogen.Field, reason ...interface{}) string {
	return errIdxCause(field, "", "nil", reason...)
}

func errCause(field protogen.Field, cause string, reason ...interface{}) string {
	return errIdxCause(field, "", cause, reason...)
}

func errIdx(field protogen.Field, idx string, reason ...interface{}) string {
	return errIdxCause(field, idx, "nil", reason...)
}

func lit(x interface{}) string {
	val := reflect.ValueOf(x)

	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.String:
		return fmt.Sprintf("%q", x)
	case reflect.Uint8:
		return fmt.Sprintf("0x%X", x)
	case reflect.Slice:
		els := make([]string, val.Len())
		for i, l := 0, val.Len(); i < l; i++ {
			els[i] = lit(val.Index(i).Interface())
		}
		return fmt.Sprintf("%T{%s}", val.Interface(), strings.Join(els, ", "))
	default:
		return fmt.Sprint(x)
	}
}

func isBytes(f interface{ Kind() protoreflect.Kind }) bool {
	return f.Kind() == protoreflect.BytesKind
}

func byteStr(x []byte) string {
	elms := make([]string, len(x))
	for i, b := range x {
		elms[i] = fmt.Sprintf(`\x%X`, b)
	}

	return fmt.Sprintf(`"%s"`, strings.Join(elms, ""))
}

func oneOfTypeName(f protogen.Field) string {
	if f.Oneof != nil {
		out := f.GoName
		parent := f.Parent.GoIdent.GoName
		return parent + "_" + out
	}
	switch f.Desc.Kind() {
	case protoreflect.BytesKind:
		return "string"
	case protoreflect.MessageKind:
		return f.Desc.Kind().String()
	case protoreflect.EnumKind:
		out := f.GoName
		parent := f.Parent.GoIdent.GoName

		return parent + "_" + out
	default:
		// Use Value() to strip any potential pointer type.
		t := f.Desc.Kind().String()
		if t == "float" {
			return "float32"
		} else if t == "double" {
			return "float64"
		}
	}
	return f.Desc.Kind().String()
	//return pgsgo.TypeName(fns.OneofOption(f)).Pointer()
}

func MapTypeName(field *protogen.Field, desc protoreflect.FieldDescriptor, rule *validate.FieldRules) ctx.FieldCtx {
	rType, rIns, mRule, wrap := ctx.ResolveRules(desc, rule)
	elem := ctx.FieldCtx{
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
	return elem
}

func inType(f protogen.Field, x interface{}) string {
	switch f.Desc.Kind() {
	case protoreflect.BytesKind:
		return "string"
	case protoreflect.MessageKind:
		switch x.(type) {
		case []*durationpb.Duration:
			return "time.Duration"
		case []string:
			return "string"
		default:
			return f.Desc.Kind().String()
			//return pgsgo.TypeName(fmt.Sprintf("%T", x)).Element().String()
		}
	case protoreflect.EnumKind:
		fullname := string(f.Enum.Desc.FullName())
		if pkgName := getOtherPkgName(fullname); pkgName != "" {
			return pkgName + "." + f.Enum.GoIdent.GoName
		}

		return f.Enum.GoIdent.GoName
	default:
		// Use Value() to strip any potential pointer type.
		t := f.Desc.Kind().String()
		switch t {
		case "float":
			return "float32"
		case "double":
			return "float64"
		case "sint32":
			return "int32"
		case "sint64":
			return "int64"
		case "fixed32":
			return "uint32"
		case "fixed64":
			return "uint64"
		case "sfixed32":
			return "int32"
		case "sfixed64":
			return "int64"

		default:
			return t
		}
	}
}

func inKey(f protogen.Field, x interface{}) string {
	switch f.Desc.Kind() {
	case protoreflect.BytesKind:
		return byteStr(x.([]byte))
	case protoreflect.MessageKind:
		switch x := x.(type) {
		case *durationpb.Duration:
			dur := x.AsDuration()
			return lit(int64(dur))
		default:
			return lit(x)
		}
	default:
		return lit(x)
	}
}

func inKind(kind protoreflect.Kind) string {
	switch kind {
	case protoreflect.BytesKind:
		return "string"
	case protoreflect.Sint64Kind:
		return "int64"
	case protoreflect.Sint32Kind:
		return "int32"
	case protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.FloatKind:
		return "float32"
	case protoreflect.DoubleKind:
		return "float64"
	case protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.Fixed32Kind:
		return "uint32"
	default:
		return kind.String()
	}
}

func durLit(dur *durationpb.Duration) string {
	return fmt.Sprintf(
		"time.Duration(%d * time.Second + %d * time.Nanosecond)",
		dur.GetSeconds(), dur.GetNanos())
}

func durStr(dur *durationpb.Duration) string {
	d := dur.AsDuration()
	return d.String()
}

func durGt(a, b *durationpb.Duration) bool {
	ad := a.AsDuration()
	bd := b.AsDuration()

	return ad > bd
}

func tsLit(ts *timestamppb.Timestamp) string {
	return fmt.Sprintf(
		"time.Unix(%d, %d)",
		ts.GetSeconds(), ts.GetNanos(),
	)
}

func tsGt(a, b *timestamppb.Timestamp) bool {
	at := a.AsTime()
	bt := b.AsTime()

	return bt.Before(at)
}

func tsStr(ts *timestamppb.Timestamp) string {
	t := ts.AsTime()
	return t.String()
}

func externalEnums(file protoreflect.FileDescriptor) []protoreflect.EnumDescriptor {
	var out []protoreflect.EnumDescriptor
	msgDescriper := file.Messages()
	for i := 0; i < msgDescriper.Len(); i++ {
		msg := msgDescriper.Get(i)
		fieldsDescriptor := msg.Fields()
		for j := 0; j < fieldsDescriptor.Len(); j++ {
			field := fieldsDescriptor.Get(j)
			var en protoreflect.EnumDescriptor
			if field.Kind() == protoreflect.EnumKind {
				en = field.Enum()
			}
			if field.IsList() {
				en = field.Enum()
			}
			if en != nil && en.ParentFile().FullName() != msg.ParentFile().FullName() {
				out = append(out, en)
			}
		}
	}

	return out
}

func enumName(enum protoreflect.EnumDescriptor, returnPkg bool) string {

	out := string(enum.Name())
	parent := enum.Parent()
	for {
		message, ok := parent.(protoreflect.MessageDescriptor)
		if ok {
			out = string(message.Name()) + "_" + out
			parent = message.Parent()
		} else {
			if returnPkg {
				if pkgName, ok := extraPkg[string(parent.FullName())]; ok {
					return pkgName + "." + out
				}
			}
			return out
		}
	}
}

type NormalizedEnum struct {
	PkgFullname string
	Name        string
}

func enumPackages(enums []protoreflect.EnumDescriptor) map[string]NormalizedEnum {
	out := make(map[string]NormalizedEnum, len(enums))

	nameCollision := map[string]int{
		"bytes":   0,
		"errors":  0,
		"fmt":     0,
		"net":     0,
		"mail":    0,
		"url":     0,
		"regexp":  0,
		"sort":    0,
		"strings": 0,
		"time":    0,
		"utf8":    0,
		"anypb":   0,
	}
	nameNormalized := make(map[string]struct{})

	for _, en := range enums {
		// TODO
		pkgName := PackageName(en)
		enImportPath := string(en.Parent().FullName())
		if _, ok := nameNormalized[pkgName]; ok {
			continue
		}

		if collision, ok := nameCollision[pkgName]; ok {
			nameCollision[pkgName] = collision + 1
			pkgName = pkgName + string(strconv.Itoa(nameCollision[pkgName]))
		}

		nameNormalized[enImportPath] = struct{}{}
		out[pkgName] = NormalizedEnum{
			Name:        enumName(en, false),
			PkgFullname: enImportPath,
		}

	}

	return out
}

func PackageName(enum protoreflect.EnumDescriptor) string {
	parent := enum.Parent()
	for {
		file, ok := parent.(protoreflect.FileDescriptor)
		if !ok {
			parent = parent.Parent()
			continue
		}
		return string(file.Name())
	}
}

var stdPkg = map[string]int{
	"bytes":   0,
	"errors":  0,
	"fmt":     0,
	"net":     0,
	"mail":    0,
	"url":     0,
	"regexp":  0,
	"sort":    0,
	"strings": 0,
	"time":    0,
	"utf8":    0,
	"anypb":   0,
}

// map[pkgPath]pkgName
var extraPkg = map[string]string{}

func GetImports(file *protogen.File) map[string]string {
	result := make(map[string]string)
	imports := file.Desc.Imports()
	if imports.Len() == 0 {
		return map[string]string{}
	}

	for i := 0; i < imports.Len(); i++ {
		imp := imports.Get(i)
		fp, ok := imp.Options().(*descriptorpb.FileOptions)
		if !ok {
			continue
		}
		pkgName := string(imp.Package().Name())
		if pkgName == "validate" || pkgName == "protobuf" {
			continue
		}

		if fp.GoPackage != nil {
			if cnt, ok1 := stdPkg[pkgName]; ok1 {
				pkgName = fmt.Sprintf("%s%d", pkgName, cnt+1)
				stdPkg[pkgName] = cnt + 1
			}
			pkgPath := strings.SplitN(*fp.GoPackage, ";", 2)[0]
			result[pkgName] = pkgPath
			extraPkg[string(imp.FullName())] = pkgName
		}
	}
	return result
}

func OnType(field *protogen.Field, typ string) string {
	var fullname string
	switch field.Desc.Kind() {
	case protoreflect.EnumKind:
		fullname = string(field.Desc.Enum().FullName())
	default:
	}
	// pkgPath := field.GoIdent.GoImportPath.String()
	if pkgName := getOtherPkgName(fullname); pkgName != "" {
		return pkgName + "." + typ
	}
	return typ
}

func getOtherPkgName(fullName string) (pkgName string) {
	for k, v := range extraPkg {
		if strings.HasPrefix(fullName, k) {
			return v
		}
	}
	return
}

func GetElemRule(field protogen.Field, rule proto.Message) (rules *validate.FieldRules, err error) {
	switch r := rule.(type) {
	case *validate.MapRules:
		rules = r.GetValues()
	case *validate.RepeatedRules:
		rules = r.GetItems()
	default:
		err = fmt.Errorf("cannot get Elem from %s", field.GoName)
		return
	}
	return
}

func SetAccessor(ctx ctx.FieldCtx, def string) ctx.FieldCtx {
	ctx.Accessor = def
	return ctx
}
