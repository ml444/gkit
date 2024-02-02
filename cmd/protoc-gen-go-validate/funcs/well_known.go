package funcs

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/v"
)

type WellKnown string

const (
	Email    WellKnown = "email"
	Hostname WellKnown = "hostname"
	UUID     WellKnown = "uuid"
)

func FileNeeds(f protogen.File, wk WellKnown) bool {
	for _, msg := range f.Messages {
		needed := Needs(msg.Desc, wk)
		if needed {
			return true
		}
	}

	return false
}

// Needs returns true if a well-known string validator is needed for this
// message.
func Needs(m protoreflect.MessageDescriptor, wk WellKnown) bool {
	fields := m.Fields()
	for i := 0; i < fields.Len(); i++ {
		f := fields.Get(i)
		//var rules v.FieldRules
		rules, ok := proto.GetExtension(f.Options(), v.E_Rules).(*v.FieldRules)
		if !ok {
			continue
		}
		switch {
		case f.IsList() && f.Kind() == protoreflect.StringKind:
			if strRulesNeeds(rules.GetRepeated().GetItems().GetString_(), wk) {
				return true
			}
		case f.IsMap():
			if f.MapKey().Kind() == protoreflect.StringKind &&
				strRulesNeeds(rules.GetMap().GetKeys().GetString_(), wk) {
				return true
			}
			if f.MapValue().Kind() == protoreflect.StringKind &&
				strRulesNeeds(rules.GetMap().GetValues().GetString_(), wk) {
				return true
			}
		case f.Kind() == protoreflect.StringKind:
			if strRulesNeeds(rules.GetString_(), wk) {
				return true
			}
			//case f.Desc.Kind() == protoreflect.StringKind && f.Type().IsEmbed() && f.Type().Embed().WellKnownType() == pgs.StringValueWKT:
			//	if strRulesNeeds(rules.GetString_(), wk) {
			//		return true
			//	}
		}
	}

	return false
}

func strRulesNeeds(rules *v.StringRules, wk WellKnown) bool {
	switch wk {
	case Email:
		if rules.GetEmail() {
			return true
		}
	case Hostname:
		if rules.GetHostname() || rules.GetAddress() {
			return true
		}
	case UUID:
		if rules.GetUuid() {
			return true
		}
	}

	return false
}
