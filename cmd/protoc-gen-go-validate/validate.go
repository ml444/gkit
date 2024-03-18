package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"google.golang.org/protobuf/reflect/protoreflect"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/ctx"
	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/v"
)

//go:embed validate.tmpl
var validateTemplate string

const Version = "1.0.0"
const deprecationComment = "// Deprecated: Do not use."

func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Messages) == 0 {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_validate.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-go-validate. DO NOT EDIT.")
	g.P(fmt.Sprintf("// - protoc-gen-go-validate %s", Version))
	g.P("// - protoc             ", protocVersion(gen))
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()

	var err error
	//buf := new(bytes.Buffer)
	//tmpl, err = tmpl.Parse(templates.FileTmpl)
	//if err != nil {
	//	panic(err.Error())
	//}
	//err = tmpl.Execute(buf, file)
	//if err != nil {
	//	panic(err.Error())
	//}
	//g.P(buf.String())
	g.P()
	fileDir := string(file.GoImportPath)
	isRelative := isRelativePath(gen.Request.GetParameter())
	if isRelative {
		// TODO: --go-validate_out=paths=source_relative:./go
		fileDir, _ = filepath.Split(file.GeneratedFilenamePrefix)
	}
	redeclaredList, err := FindRedeclaredInPkg(fileDir, file.GeneratedFilenamePrefix)
	if err != nil {
		panic(err.Error())
	}
	//if !redeclaredList {
	//	g.P(templates.CommonDefTmpl)
	//}
	generateFileContent(file, g, redeclaredList)
	return g
}

func generateFileContent(file *protogen.File, g *protogen.GeneratedFile, redeclaredList []string) {
	var err error
	if len(file.Messages) == 0 {
		return
	}
	tmpl := template.New("file")
	Register(tmpl)
	tmpl, err = tmpl.New("validate").Parse(strings.TrimSpace(validateTemplate))
	if err != nil {
		panic(err.Error())
	}

	validateCtx := &ctx.ValidateCtx{}
	validateCtx.ErrCodeBegin = getErrCodeBegin(file)

	needWKn := &ctx.NeedWellKnown{}
	importMap := make(map[string]string)
	for _, message := range file.Messages {
		need, msgCtx, imports := genMessage(message, needWKn)
		if !need {
			continue
		}
		validateCtx.Messages = append(validateCtx.Messages, msgCtx)
		for _, imp := range imports {
			importMap[imp.Path] = imp.Alias
		}
	}
	// TODO sort import
	for path, alias := range importMap {
		validateCtx.Imports = append(validateCtx.Imports, &ctx.ImportCtx{Alias: alias, Path: path})
	}
	redeclaredMap := make(map[string]bool)
	for _, redeclared := range redeclaredList {
		redeclaredMap[redeclared] = true
	}
	if needWKn.Email && redeclaredMap["_validateEmail"] {
		needWKn.Email = false
	}
	if needWKn.Hostname && redeclaredMap["_validateHostname"] {
		needWKn.Hostname = false
	}
	if needWKn.UUID && redeclaredMap["_validateUuid"] {
		needWKn.UUID = false
	}
	validateCtx.NeedWellKnow = needWKn
	if !redeclaredMap["ValidationError"] {
		validateCtx.Imports = append(validateCtx.Imports, &ctx.ImportCtx{Alias: "", Path: "github.com/ml444/gkit/errorx"})
		validateCtx.NeedCommon = true
	}

	buf := new(bytes.Buffer)
	err = tmpl.Lookup("validate").Execute(buf, validateCtx)
	if err != nil {
		panic(err.Error())
	}
	_, err = g.Write(buf.Bytes())
	if err != nil {
		panic(err.Error())
	}
}

func getErrCodeBegin(file *protogen.File) int32 {
	var errCodeBegin int32
	for _, enum := range file.Enums {
		if errcode, ok := proto.GetExtension(enum.Desc.Options(), v.E_LowerBound).(int32); ok && errcode != 0 {
			errCodeBegin = errcode
		} else {
			if len(enum.Values) > 0 {
				firstValue := enum.Values[0]
				errCodeBegin = int32(firstValue.Desc.Number())
			}
		}
	}

	return errCodeBegin
}

func genMessage(message *protogen.Message, needWKn *ctx.NeedWellKnown) (bool, *ctx.MessageCtx, []*ctx.ImportCtx) {
	var imports []*ctx.ImportCtx
	msgCtx := &ctx.MessageCtx{
		Desc:           message.Desc,
		TypeName:       message.GoIdent.GoName,
		NonOneOfFields: []*ctx.FieldCtx{},
		RealOneOfs:     map[string]*ctx.OneOfField{},
		OptionalFields: []*ctx.FieldCtx{},
	}
	disabled, ok := proto.GetExtension(message.Desc.Options(), v.E_Disabled).(bool)
	if ok {
		msgCtx.Disabled = disabled
	}
	ignored, ok := proto.GetExtension(message.Desc.Options(), v.E_Ignored).(bool)
	if ok {
		msgCtx.Ignored = ignored
	}
	var needGen bool
	for _, field := range message.Fields {
		rule, ok := proto.GetExtension(field.Desc.Options(), v.E_Rules).(*v.FieldRules)
		if ok && rule != nil {
			needWellKnown(field.Desc, rule, needWKn)
			needGen = true
			ruleType, ruleIns, messageRule, wrapped := ctx.ResolveRules(field.Desc, rule)
			fieldCtx := ctx.FieldCtx{
				Desc:     field.Desc,
				Field:    field,
				Rules:    ruleIns,
				Name:     field.GoName,
				Type:     field.Desc.Kind().String(),
				TmplName: ruleType,
				Err:      nil,
			}
			importCtx := fieldCtx.ImpCtx()
			if importCtx != nil {
				imports = append(imports, importCtx)
			}
			if rule.Errcode != nil {
				fieldCtx.ErrCode = *rule.Errcode
			}
			if wrapped {
				fieldCtx.Wrap = ruleType
				fieldCtx.TmplName = "wrapper"
			}
			if field.Enum != nil {
				fieldCtx.Type = field.Enum.GoIdent.GoName
			} else if field.Message != nil {
				fieldCtx.Type = field.Message.GoIdent.GoName
			}

			if messageRule != nil {
				if messageRule.Required != nil {
					fieldCtx.Required = *messageRule.Required
				}
				if messageRule.Skip != nil {
					fieldCtx.Skip = *messageRule.Skip
				}
			}
			msgCtx.Fields = append(msgCtx.Fields, &fieldCtx)
			if field.Oneof != nil {
				if isOptional(field) {
					msgCtx.OptionalFields = append(msgCtx.OptionalFields, &fieldCtx)
				} else {
					handleOneOfs(field, &fieldCtx, msgCtx)
				}
			} else {
				msgCtx.NonOneOfFields = append(msgCtx.NonOneOfFields, &fieldCtx)
			}
		} else {
			if field.Oneof != nil {
				if isOptional(field) {
					continue
				}
				fieldCtx := ctx.FieldCtx{
					Desc:     field.Desc,
					Field:    field,
					Rules:    nil,
					Name:     field.GoName,
					Type:     field.Desc.Kind().String(),
					TmplName: "none",
					Err:      nil,
				}
				handleOneOfs(field, &fieldCtx, msgCtx)
			}

			//if field.Desc.Kind() == protoreflect.MessageKind {
			//	fieldCtx := ctx.FieldCtx{
			//		Desc:  field.Desc,
			//		Rules: &v.MessageRules{},
			//		Field: field,
			//		Name:  field.GoName,
			//		Type:  field.Desc.Kind().String(),
			//		//Required: *messageRule.Required,
			//		//Skip:     *messageRule.Skip,
			//		TmplName: "message",
			//		Err:      nil,
			//	}
			//	msgCtx.NonOneOfFields = append(msgCtx.NonOneOfFields, &fieldCtx)
			//}
		}
	}

	for _, msg := range message.Messages {
		need, subMsgCtx, subImports := genMessage(msg, needWKn)
		if !need {
			continue
		}
		msgCtx.SubMessageCtxs = append(msgCtx.SubMessageCtxs, subMsgCtx)
		imports = append(imports, subImports...)
	}
	return needGen, msgCtx, imports
	//if needGen {
	//	buf := new(bytes.Buffer)
	//	err := tmpl.Lookup("validate").Execute(buf, msgCtx)
	//	if err != nil {
	//		panic(err.Error())
	//	}
	//	_, err = g.Write(buf.Bytes())
	//	if err != nil {
	//		panic(err.Error())
	//	}
	//	for _, msg := range message.Messages {
	//		genMessage(tmpl, file, g, msg)
	//	}
	//}
}

func isOptional(field *protogen.Field) bool {
	if field.Desc.Cardinality() != protoreflect.Optional {
		return false
	}
	if fieldName := field.Desc.JSONName(); unicode.IsUpper(rune(fieldName[0])) {
		if strings.HasPrefix(string(field.Oneof.Desc.Name()), "X_") {
			return true
		}
	} else if strings.HasPrefix(string(field.Oneof.Desc.Name()), "_") {
		return true
	}
	return false
}

func needWellKnown(f protoreflect.FieldDescriptor, rules *v.FieldRules, wk *ctx.NeedWellKnown) {
	var strRules *v.StringRules
	switch {
	case f.IsList() && f.Kind() == protoreflect.StringKind:
		strRules = rules.GetRepeated().GetItems().GetString_()
	case f.IsMap():
		if f.MapKey().Kind() == protoreflect.StringKind {
			strRules = rules.GetMap().GetKeys().GetString_()
		}
		if f.MapValue().Kind() == protoreflect.StringKind {
			strRules = rules.GetMap().GetValues().GetString_()
		}
	case f.Kind() == protoreflect.StringKind:
		strRules = rules.GetString_()
	}
	if strRules != nil {
		if strRules.GetEmail() {
			wk.Email = true
		}
		if strRules.GetHostname() || strRules.GetAddress() {
			wk.Hostname = true
		}
		if strRules.GetUuid() {
			wk.UUID = true
		}
	}
}

func handleOneOfs(field *protogen.Field, fieldCtx *ctx.FieldCtx, msgData *ctx.MessageCtx) {
	oneOf, ok := msgData.RealOneOfs[field.Oneof.GoName]
	if !ok {
		oneOf = &ctx.OneOfField{
			Fields: nil,
			Field:  field,
			Name:   field.Oneof.GoName,
			Type:   string(field.Oneof.Desc.Name()),
			//Required: *messageRule.Required,
			//Skip:     *messageRule.Skip,
			//TmplName: ruleType,
		}
	}
	required, ok := proto.GetExtension(field.Oneof.Desc.Options(), v.E_Required).(bool)
	if ok {
		oneOf.Required = required
	}
	oneOf.Fields = append(oneOf.Fields, fieldCtx)
	msgData.RealOneOfs[oneOf.Name] = oneOf
}

func protocVersion(gen *protogen.Plugin) string {
	ver := gen.Request.GetCompilerVersion()
	if ver == nil {
		return "(unknown)"
	}
	var suffix string
	if s := ver.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", ver.GetMajor(), ver.GetMinor(), ver.GetPatch(), suffix)
}
