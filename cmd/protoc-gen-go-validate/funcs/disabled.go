package funcs

import (
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"

	"github.com/ml444/gkit/cmd/protoc-gen-go-validate/validate"
)

// Disabled returns true if validations are disabled for msg
func Disabled(msg protogen.Message) (disabled bool) {
	disabled = proto.GetExtension(msg.Desc.Options(), validate.E_Disabled).(bool)
	return
}

// Ignored returns true if validations aren't to be generated for msg
func Ignored(msg protogen.Message) (ignored bool) {
	return proto.GetExtension(msg.Desc.Options(), validate.E_Ignored).(bool)
}

// RequiredOneOf returns true if the oneof field requires a field to be set
func RequiredOneOf(oo protogen.Oneof) (required bool) {
	return proto.GetExtension(oo.Desc.Options(), validate.E_Required).(bool)
}
