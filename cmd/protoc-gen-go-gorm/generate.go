package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/ml444/gkit/cmd/protoc-gen-go-gorm/orm"
	"github.com/ml444/gkit/cmd/protoc-gen-go-gorm/templates"
)

const (
	release            = "v1.0.0"
	deprecationComment = "// Deprecated: Do not use."
)

//go:embed orm_pb.tmpl
var serializerTemplate string

func protocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}

func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Messages) == 0 {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_orm.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-go-gorm. DO NOT EDIT.")
	g.P("// versions:")
	g.P(fmt.Sprintf("// - protoc-gen-go-gorm %s", release))
	g.P("// - protoc             ", protocVersion(gen))
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
	g.P("package ", file.GoPackageName)
	err := genContent(file, g)
	if err != nil {
		gen.Error(err)
	}
	return g
}

func Render(tpl *template.Template) func(tmplName string, data interface{}) (string, error) {
	return func(tmplName string, data interface{}) (string, error) {
		var b bytes.Buffer
		err := tpl.ExecuteTemplate(&b, tmplName, data)
		return b.String(), err
	}
}

func genContent(file *protogen.File, g *protogen.GeneratedFile) error {
	var err error
	tmpl := template.New("serializer")
	funcMap := template.FuncMap{
		"render":             Render(tmpl),
		"SnakeToCamel":       orm.SnakeToCamel,
		"JoinStringsByCamel": orm.JoinStringsByCamel,
	}
	tmpl.Funcs(funcMap)
	tmpl, err = tmpl.Parse(strings.TrimSpace(serializerTemplate))
	if err != nil {
		return err
	}
	var messages []*orm.MessageDesc
	parseMessages(g, file.Messages, &messages)
	if len(messages) == 0 {
		return nil
	}

	importMap := map[string]bool{}
	tmplMap := map[string]string{}
	commonMap := map[string]string{}
	for _, m := range messages {
		for _, imp := range m.Imports {
			importMap[imp] = true
		}
		for _, f := range m.SerializeFields {
			tmplMap[f.SerializerName] = f.Tmpl
		}
		for key, str := range m.UtilMap {
			commonMap[key] = str
		}
	}
	var imports []string
	for key := range importMap {
		imports = append(imports, key)
	}
	var commons []string
	for _, v := range commonMap {
		commons = append(commons, v)
	}
	sort.Slice(imports, func(i, j int) bool {
		return imports[i] < imports[j]
	})
	fd := &orm.FileDesc{
		PackageName: string(file.GoPackageName),
		Imports:     imports,
		Messages:    messages,
		Commons:     commons,
	}

	// register template
	for name, tpl := range tmplMap {
		template.Must(tmpl.New(name).Parse(tpl))
	}
	return tmpl.Execute(g, fd)
}

func parseMessages(g *protogen.GeneratedFile, messages []*protogen.Message, result *[]*orm.MessageDesc) {
	for _, message := range messages {
		if message.Desc.Options().(*descriptorpb.MessageOptions).GetDeprecated() {
			g.P("//")
			g.P(deprecationComment)
		}
		if message.Desc.Options().(*descriptorpb.MessageOptions) == nil {
			continue
		}
		if enable, ok := proto.GetExtension(message.Desc.Options(), orm.E_Enable).(bool); !ok || !enable {
			continue
		}
		msgDesc := orm.MessageDesc{
			Name:    message.GoIdent.GoName,
			Opts:    &orm.MessageOpts{},
			Fields:  make([]*orm.ORMField, 0),
			UtilMap: map[string]string{},
		}
		// parse message options
		if tableName, ok := proto.GetExtension(message.Desc.Options(), orm.E_TableName).(string); ok && tableName != "" {
			msgDesc.Opts.TableName = tableName
		}
		if indexClauses, ok := proto.GetExtension(message.Desc.Options(), orm.E_IndexClauses).([]*orm.IndexClause); ok && indexClauses != nil {
			msgDesc.Opts.IndexClauses = indexClauses
			msgDesc.Imports = append(msgDesc.Imports, "gorm.io/gorm/clause", "gorm.io/hints")
		}
		//if forceIdx, ok := proto.GetExtension(message.Desc.Options(), orm.E_ForceIndex).(string); ok && forceIdx != "" {
		//	msgDesc.Opts.ForceIndex = forceIdx
		//}
		//if ignoreIdx, ok := proto.GetExtension(message.Desc.Options(), orm.E_IgnoreIndex).(string); ok && ignoreIdx != "" {
		//	msgDesc.Opts.IgnoreIndex = ignoreIdx
		//}
		// parse message fields
		for _, field := range message.Fields {
			tags, ok := proto.GetExtension(field.Desc.Options(), orm.E_Tags).(*orm.ORMTags)
			if !ok || tags == nil {
				continue
			}
			forceORM, tagStr := orm.JoinORMTags(tags)
			fieldName := field.GoName
			oldType := goType(field)
			ormField := &orm.ORMField{
				FieldName: fieldName,
				NewType:   oldType,
				OldType:   oldType,
				ORMTag:    orm.JoinTags(string(field.Desc.Name()), tagStr),
			}
			msgDesc.ForceORM = forceORM
			msgDesc.Fields = append(msgDesc.Fields, ormField)
			if len(NeedGenerateFunctionFields) > 0 {
				if ok1 := NeedGenerateFunctionFields[fieldName]; ok1 {
					msgDesc.NeedGenFuncFields = append(msgDesc.NeedGenFuncFields, ormField)
				}
			}

			typ := ""
			if tags.Type != nil && *tags.Type != "" {
				typ = *tags.Type
			}
			if isBaseType(field) && typ == "" {
				continue
			}

			sType := serializeType(field)
			sd := orm.SerializeDesc{
				SerializerName:     strings.ToLower(typ),
				SerializerTypeName: sType,
				FieldType:          ormField.NewType,
			}

			if tags.IgnoreMigration != nil && *tags.IgnoreMigration {
				sd.IsIgnore = true
			}
			if tags.IgnoreAll != nil && *tags.IgnoreAll {
				sd.IsIgnore = true
			}
			var imports []string
			needGenSerializer := true
			switch strings.ToLower(typ) {
			case "json", "text", "mediumtext", "longtext":
				if field.Desc.Kind() == protoreflect.StringKind {
					needGenSerializer = false
					break
				}
				if field.Desc.Cardinality() == protoreflect.Repeated || field.Desc.IsMap() {
					sd.SerializerName = "special_json"
					imports = templates.SpecialJsonImports
					sd.Tmpl = templates.SpecialJsonTmpl
					break
				}
				msgDesc.UtilMap["jsonMarshal"] = templates.JsonUtils
				imports = templates.JsonImports
				sd.Tmpl = templates.JsonTmpl
			case "bytes", "blob", "mediumblob", "longblob":
				msgDesc.UtilMap["bytesMarshal"] = templates.BytesUtils
				if field.Desc.Kind() == protoreflect.BytesKind ||
					field.Desc.Cardinality() == protoreflect.Repeated ||
					field.Desc.IsMap() {
					sd.SerializerName = "special_bytes"
					imports = templates.SpecialBytesImports
					sd.Tmpl = templates.SpecialBytesTmpl
					break
				}
				sd.SerializerName = "bytes"
				imports = templates.BytesImports
				sd.Tmpl = templates.BytesTmpl
			case "date", "time", "datetime":
				msgDesc.UtilMap["datetime"] = templates.DateUtils
				imports = templates.DateImports
				sd.Tmpl = templates.DateTmpl
			case "":
				if field.Desc.Cardinality() == protoreflect.Repeated || field.Desc.IsMap() {
					sd.SerializerName = "special_json"
					imports = templates.SpecialJsonImports
					sd.Tmpl = templates.SpecialJsonTmpl
				} else if field.Desc.Kind() == protoreflect.BytesKind {
					sd.SerializerName = "special_bytes"
					imports = templates.SpecialBytesImports
					sd.Tmpl = templates.SpecialBytesTmpl
				} else if field.Desc.Kind() == protoreflect.MessageKind || field.Desc.Kind() == protoreflect.GroupKind {
					sd.SerializerName = "special_json"
					imports = templates.SpecialJsonImports
					sd.Tmpl = templates.SpecialJsonTmpl
				} else {
					needGenSerializer = false
				}
			default:
				needGenSerializer = false
				//if field.Desc.Kind() == protoreflect.BytesKind {
				//	sd.SerializerName = "special_bytes"
				//	sd.FieldType = goType(field)
				//	imports = templates.SpecialBytesImports
				//	sd.Tmpl = templates.SpecialBytesTmpl
				//} else {
				//	sd.SerializerName = "special_json"
				//	sd.FieldType = goType(field)
				//	imports = templates.SpecialJsonImports
				//	sd.Tmpl = templates.SpecialJsonTmpl
				//}
			}
			if needGenSerializer {
				// replace field type
				if field.Desc.HasPresence() {
					ormField.NewType = "*" + sType
				} else {
					ormField.NewType = sType
				}

				msgDesc.SerializeFields = append(msgDesc.SerializeFields, &sd)
				msgDesc.Imports = append(msgDesc.Imports, imports...)
			}
		}
		*result = append(*result, &msgDesc)
		if len(message.Messages) > 0 {
			parseMessages(g, message.Messages, result)
		}
	}
}

func goType(field *protogen.Field) string {
	typ := fieldDescToType(field.Desc)
	if typ != "" {
		return typ
	}
	return specialType(field)
}

func serializeType(field *protogen.Field) string {
	if field.Desc.Cardinality() == protoreflect.Repeated || field.Desc.IsMap() {
		return field.GoIdent.GoName
	}
	typ := specialType(field)
	if typ != "" {
		return strings.TrimPrefix(typ, "*")
	}
	return field.GoIdent.GoName
}

func isBaseType(field *protogen.Field) bool {
	if field.Desc.Cardinality() == protoreflect.Repeated || field.Desc.IsMap() {
		return false
	}
	switch field.Desc.Kind() {
	case protoreflect.BytesKind, protoreflect.MessageKind, protoreflect.GroupKind:
		return false
	}
	return true
}

func kindToGoType(kind protoreflect.Kind) string {
	switch kind {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float32"
	case protoreflect.DoubleKind:
		return "float64"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.BytesKind:
		return "[]byte"
	// case protoreflect.MessageKind:
	//	return "struct"
	// case protoreflect.GroupKind:
	//	return "struct"
	// case protoreflect.EnumKind:
	//	return "enum"
	default:
		return ""
	}
}

func fieldDescToType(fieldDesc protoreflect.FieldDescriptor) string {
	switch fieldDesc.Cardinality() {
	case protoreflect.Repeated:
		typ := kindToGoType(fieldDesc.Kind())
		if typ != "" {
			return "[]" + typ
		}
	case protoreflect.Optional:
		typ := kindToGoType(fieldDesc.Kind())
		if typ != "" {
			if fieldDesc.HasPresence() {
				return "*" + typ
			}
			return typ
		}
	}
	if fieldDesc.IsMap() {
		key := kindToGoType(fieldDesc.MapKey().Kind())
		value := kindToGoType(fieldDesc.MapValue().Kind())
		if key == "" || value == "" {
			return ""
		}
		return "map[" + key + "]" + value
	}
	return ""
}

func specialType(field *protogen.Field) string {
	switch field.Desc.Kind() {
	case protoreflect.MessageKind, protoreflect.GroupKind:
		if field.Desc.IsMap() {
			key := fieldDescToType(field.Desc.MapKey())
			valueDesc := field.Desc.MapValue()
			value := fieldDescToType(valueDesc)
			if value == "" {
				if valueDesc.Kind() == protoreflect.MessageKind || valueDesc.Kind() == protoreflect.GroupKind {
					value = typeName(valueDesc.Message())
					if valueDesc.HasPresence() {
						value = "*" + value
					}
				} else if valueDesc.Kind() == protoreflect.EnumKind {
					value = typeName(valueDesc.Enum())
				}
			}
			return "map[" + key + "]" + value
		} else if field.Desc.Cardinality() == protoreflect.Repeated {
			typ := typeName(field.Desc.Message())
			return "[]*" + typ
		}

		typ := field.Message.GoIdent.GoName
		if field.Desc.HasPresence() {
			return "*" + typ
		}
		return typ
	case protoreflect.EnumKind:
		typ := field.Enum.GoIdent.GoName
		if field.Desc.HasPresence() {
			return "*" + typ
		}
		return typ
	}

	return ""
}

func typeName(fieldDesc protoreflect.Descriptor) string {
	name := strings.TrimPrefix(string(fieldDesc.FullName()), string(fieldDesc.ParentFile().Package())+".")
	return strings.ReplaceAll(name, ".", "_")
}
