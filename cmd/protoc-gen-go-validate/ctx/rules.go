package ctx

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/validate"
)

func isEmbed(desc protoreflect.FieldDescriptor) bool {
	if desc.Kind() == protoreflect.MessageKind {
		return true
	}
	return false
}

func ResolveRules(desc protoreflect.FieldDescriptor, rules *validate.FieldRules) (ruleType string, rule proto.Message, messageRule *validate.MessageRules, wrapped bool) {
	if rules == nil {
		rules = &validate.FieldRules{}
	}
	switch r := rules.GetType().(type) {
	case *validate.FieldRules_Float:
		ruleType, rule, wrapped = "float", r.Float, isEmbed(desc) // typ.IsEmbed()
	case *validate.FieldRules_Double:
		ruleType, rule, wrapped = "double", r.Double, isEmbed(desc) // typ.IsEmbed()
	case *validate.FieldRules_Int32:
		ruleType, rule, wrapped = "int32", r.Int32, isEmbed(desc) // typ.IsEmbed()
	case *validate.FieldRules_Int64:
		ruleType, rule, wrapped = "int64", r.Int64, isEmbed(desc) // typ.IsEmbed()
	case *validate.FieldRules_Uint32:
		ruleType, rule, wrapped = "uint32", r.Uint32, isEmbed(desc) // typ.IsEmbed()
	case *validate.FieldRules_Uint64:
		ruleType, rule, wrapped = "uint64", r.Uint64, isEmbed(desc) // typ.IsEmbed()
	case *validate.FieldRules_Sint32:
		ruleType, rule, wrapped = "sint32", r.Sint32, false
	case *validate.FieldRules_Sint64:
		ruleType, rule, wrapped = "sint64", r.Sint64, false
	case *validate.FieldRules_Fixed32:
		ruleType, rule, wrapped = "fixed32", r.Fixed32, false
	case *validate.FieldRules_Fixed64:
		ruleType, rule, wrapped = "fixed64", r.Fixed64, false
	case *validate.FieldRules_Sfixed32:
		ruleType, rule, wrapped = "sfixed32", r.Sfixed32, false
	case *validate.FieldRules_Sfixed64:
		ruleType, rule, wrapped = "sfixed64", r.Sfixed64, false
	case *validate.FieldRules_Bool:
		ruleType, rule, wrapped = "bool", r.Bool, isEmbed(desc) // typ.IsEmbed()
	case *validate.FieldRules_String_:
		ruleType, rule, wrapped = "string", r.String_, isEmbed(desc) // typ.IsEmbed()
	case *validate.FieldRules_Bytes:
		ruleType, rule, wrapped = "bytes", r.Bytes, isEmbed(desc) // typ.IsEmbed()
	case *validate.FieldRules_Enum:
		ruleType, rule, wrapped = "enum", r.Enum, false
	case *validate.FieldRules_Repeated:
		ruleType, rule, wrapped = "repeated", r.Repeated, false
	case *validate.FieldRules_Map:
		ruleType, rule, wrapped = "map", r.Map, false
	case *validate.FieldRules_Any:
		ruleType, rule, wrapped = "any", r.Any, false
	case *validate.FieldRules_Duration:
		ruleType, rule, wrapped = "duration", r.Duration, false
	case *validate.FieldRules_Timestamp:
		ruleType, rule, wrapped = "timestamp", r.Timestamp, false
	case nil:
		if desc.IsList() {
			return "repeated", &validate.RepeatedRules{}, rules.Message, false
		} else if desc.IsMap() {
			return "map", &validate.MapRules{}, rules.Message, false
		} else if desc.Kind() == protoreflect.MessageKind {
			return "message", rules.GetMessage(), rules.GetMessage(), false
		}
		return "none", nil, nil, false
	default:
		ruleType, rule, wrapped = "error", nil, false
	}

	return ruleType, rule, rules.Message, wrapped
}
