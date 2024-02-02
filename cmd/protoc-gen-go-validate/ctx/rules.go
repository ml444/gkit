package ctx

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/v"
)

func isEmbed(desc protoreflect.FieldDescriptor) bool {
	if desc.Kind() == protoreflect.MessageKind {
		return true
	}
	return false
}

func ResolveRules(desc protoreflect.FieldDescriptor, rules *v.FieldRules) (ruleType string, rule proto.Message, messageRule *v.MessageRules, wrapped bool) {
	if rules == nil {
		rules = &v.FieldRules{}
	}
	switch r := rules.GetType().(type) {
	case *v.FieldRules_Float:
		ruleType, rule, wrapped = "float", r.Float, isEmbed(desc) // typ.IsEmbed()
	case *v.FieldRules_Double:
		ruleType, rule, wrapped = "double", r.Double, isEmbed(desc) // typ.IsEmbed()
	case *v.FieldRules_Int32:
		ruleType, rule, wrapped = "int32", r.Int32, isEmbed(desc) // typ.IsEmbed()
	case *v.FieldRules_Int64:
		ruleType, rule, wrapped = "int64", r.Int64, isEmbed(desc) // typ.IsEmbed()
	case *v.FieldRules_Uint32:
		ruleType, rule, wrapped = "uint32", r.Uint32, isEmbed(desc) // typ.IsEmbed()
	case *v.FieldRules_Uint64:
		ruleType, rule, wrapped = "uint64", r.Uint64, isEmbed(desc) // typ.IsEmbed()
	case *v.FieldRules_Sint32:
		ruleType, rule, wrapped = "sint32", r.Sint32, false
	case *v.FieldRules_Sint64:
		ruleType, rule, wrapped = "sint64", r.Sint64, false
	case *v.FieldRules_Fixed32:
		ruleType, rule, wrapped = "fixed32", r.Fixed32, false
	case *v.FieldRules_Fixed64:
		ruleType, rule, wrapped = "fixed64", r.Fixed64, false
	case *v.FieldRules_Sfixed32:
		ruleType, rule, wrapped = "sfixed32", r.Sfixed32, false
	case *v.FieldRules_Sfixed64:
		ruleType, rule, wrapped = "sfixed64", r.Sfixed64, false
	case *v.FieldRules_Bool:
		ruleType, rule, wrapped = "bool", r.Bool, isEmbed(desc) // typ.IsEmbed()
	case *v.FieldRules_String_:
		ruleType, rule, wrapped = "string", r.String_, isEmbed(desc) // typ.IsEmbed()
	case *v.FieldRules_Bytes:
		ruleType, rule, wrapped = "bytes", r.Bytes, isEmbed(desc) // typ.IsEmbed()
	case *v.FieldRules_Enum:
		ruleType, rule, wrapped = "enum", r.Enum, false
	case *v.FieldRules_Repeated:
		ruleType, rule, wrapped = "repeated", r.Repeated, false
	case *v.FieldRules_Map:
		ruleType, rule, wrapped = "map", r.Map, false
	case *v.FieldRules_Any:
		ruleType, rule, wrapped = "any", r.Any, false
	case *v.FieldRules_Duration:
		ruleType, rule, wrapped = "duration", r.Duration, false
	case *v.FieldRules_Timestamp:
		ruleType, rule, wrapped = "timestamp", r.Timestamp, false
	case nil:
		if desc.IsList() {
			return "repeated", &v.RepeatedRules{}, rules.Message, false
		} else if desc.IsMap() {
			return "map", &v.MapRules{}, rules.Message, false
		} else if desc.Kind() == protoreflect.MessageKind {
			return "message", rules.GetMessage(), rules.GetMessage(), false
		}
		return "none", nil, nil, false
	default:
		ruleType, rule, wrapped = "error", nil, false
	}

	return ruleType, rule, rules.Message, wrapped
}
