package main

import (
	"testing"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/ml444/gkit/cmd/protoc-gen-go-gorm/orm"
	"github.com/ml444/gkit/cmd/protoc-gen-go-gorm/templates"
)

func TestSortedCommons(t *testing.T) {
	got := sortedCommons(map[string]string{
		"jsonMarshal":  "json-func",
		"bytesMarshal": "bytes-func",
		"datetime":     "date-func",
	})
	want := []string{"bytes-func", "date-func", "json-func"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("sortedCommons[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestAppendImportsDedupes(t *testing.T) {
	var imports []string
	appendImports(&imports, "fmt", "fmt", "encoding/json")
	appendImports(&imports, "encoding/json", "time")
	if len(imports) != 3 {
		t.Fatalf("imports = %v, want 3 unique entries", imports)
	}
}

func TestGoTypeRepeatedEnumFromAnotherPackage(t *testing.T) {
	commonFile := &descriptorpb.FileDescriptorProto{
		Name:    proto.String("common/common.proto"),
		Package: proto.String("common"),
		Syntax:  proto.String("proto3"),
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("example.com/project/common;common")},
		EnumType: []*descriptorpb.EnumDescriptorProto{{
			Name: proto.String("Outcome"),
			Value: []*descriptorpb.EnumValueDescriptorProto{{
				Name:   proto.String("OUTCOME_UNSPECIFIED"),
				Number: proto.Int32(0),
			}},
		}},
	}
	poolFile := &descriptorpb.FileDescriptorProto{
		Name:       proto.String("pool/pool.proto"),
		Package:    proto.String("pool"),
		Syntax:     proto.String("proto3"),
		Dependency: []string{"common/common.proto"},
		Options:    &descriptorpb.FileOptions{GoPackage: proto.String("example.com/project/pool;pool")},
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: proto.String("ModelPool"),
			Field: []*descriptorpb.FieldDescriptorProto{{
				Name:     proto.String("allowed_vote_outcomes"),
				Number:   proto.Int32(19),
				Label:    descriptorpb.FieldDescriptorProto_LABEL_REPEATED.Enum(),
				Type:     descriptorpb.FieldDescriptorProto_TYPE_ENUM.Enum(),
				TypeName: proto.String(".common.Outcome"),
			}},
		}},
	}
	gen, err := (protogen.Options{}).New(&pluginpb.CodeGeneratorRequest{
		ProtoFile:      []*descriptorpb.FileDescriptorProto{commonFile, poolFile},
		FileToGenerate: []string{poolFile.GetName()},
	})
	if err != nil {
		t.Fatalf("create protogen plugin: %v", err)
	}

	var generatedPoolFile *protogen.File
	for _, file := range gen.Files {
		if file.Desc.Path() == poolFile.GetName() {
			generatedPoolFile = file
			break
		}
	}
	if generatedPoolFile == nil {
		t.Fatal("pool proto file not found")
	}
	field := generatedPoolFile.Messages[0].Fields[0]
	g := gen.NewGeneratedFile("pool_orm.pb.go", generatedPoolFile.GoImportPath)

	if got, want := goType(g, field), "[]common.Outcome"; got != want {
		t.Fatalf("goType() = %q, want %q", got, want)
	}
}

func TestDedupeSerializeFieldsAcrossMessages(t *testing.T) {
	newJSONField := func() *orm.SerializeDesc {
		return &orm.SerializeDesc{
			SerializerName:     "json",
			SerializerTypeName: "PoolTerms",
			FieldType:          "*PoolTerms",
			Tmpl:               templates.JsonTmpl,
		}
	}
	messages := []*orm.MessageDesc{
		{Name: "ModelDraft", SerializeFields: []*orm.SerializeDesc{newJSONField()}},
		{Name: "ModelPool", SerializeFields: []*orm.SerializeDesc{newJSONField()}},
	}

	dedupeSerializeFields(messages)

	if got := len(messages[0].SerializeFields); got != 1 {
		t.Fatalf("first message serializer count = %d, want 1", got)
	}
	if got := len(messages[1].SerializeFields); got != 0 {
		t.Fatalf("second message serializer count = %d, want 0", got)
	}
}
