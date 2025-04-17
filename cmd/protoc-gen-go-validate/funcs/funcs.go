package funcs

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/v"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var FileAlias string

func GetAliasName() string {
	return FileAlias
}

func Render(tpl *template.Template) func(tmplName string, data interface{}) (string, error) {
	return func(tmplName string, data interface{}) (string, error) {
		var b bytes.Buffer
		err := tpl.ExecuteTemplate(&b, tmplName, data)
		return b.String(), err
	}
}

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
	var errCode int32
	rule, ok := proto.GetExtension(field.Desc.Options(), v.E_Rules).(*v.FieldRules)
	if ok && rule.Errcode != nil {
		errCode = *rule.Errcode
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
		errCode: %d,
		%s%s
	}`,
		fmt.Sprintf("%sValidationError", FileAlias),
		fld,
		fmt.Sprint(reason...),
		errCode,
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
	// return pgsgo.TypeName(fns.OneofOption(f)).Pointer()
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
			// return pgsgo.TypeName(fmt.Sprintf("%T", x)).Element().String()
		}
	case protoreflect.EnumKind:
		fullname := string(f.Desc.Enum().FullName())
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
	for k, v := range ExtraPkg {
		if strings.HasPrefix(fullName, k) {
			return v
		}
	}
	return
}

func GetElemRule(field protogen.Field, rule proto.Message) (rules *v.FieldRules, err error) {
	switch r := rule.(type) {
	case *v.MapRules:
		rules = r.GetValues()
	case *v.RepeatedRules:
		rules = r.GetItems()
	default:
		err = fmt.Errorf("cannot get Elem from %s", field.GoName)
		return
	}
	return
}

func JoinString(ss ...string) string {
	return strings.Join(ss, "")
}
