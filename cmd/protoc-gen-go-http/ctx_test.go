package main

import (
	"reflect"
	"testing"

	"github.com/ml444/gkit/cmd/protoc-gen-go-http/pluck"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func Test_pluckFields(t *testing.T) {
	ct := "application/json"
	got := pluckFields(&pluck.RequestHeaders{ContentType: &ct})
	want := map[string]string{"Content-Type": "application/json"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("pluckFields() = %v, want %v", got, want)
	}
}

func TestIsExistField(t *testing.T) {
	msgDesc := buildUploadReqDesc(t)
	tests := []struct {
		field string
		want  bool
	}{
		{"file_info", true},
		{"file_info.file_name", true},
		{"file_info.missing", false},
		{"missing", false},
	}
	for _, tt := range tests {
		if got := isExistField(msgDesc, tt.field); got != tt.want {
			t.Errorf("isExistField(%q) = %v, want %v", tt.field, got, tt.want)
		}
	}
}

func buildUploadReqDesc(t *testing.T) protoreflect.MessageDescriptor {
	t.Helper()
	fileInfo := &descriptorpb.DescriptorProto{
		Name: protoString("FileInfo"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{Name: protoString("file_name"), Number: protoInt32(1), Type: protoFieldType(descriptorpb.FieldDescriptorProto_TYPE_STRING)},
		},
	}
	uploadReq := &descriptorpb.DescriptorProto{
		Name: protoString("UploadReq"),
		Field: []*descriptorpb.FieldDescriptorProto{
			{Name: protoString("file_info"), Number: protoInt32(1), Type: protoFieldType(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE), TypeName: protoString(".storage.FileInfo")},
		},
		NestedType: []*descriptorpb.DescriptorProto{fileInfo},
	}
	file := &descriptorpb.FileDescriptorProto{
		Name:    protoString("test.proto"),
		Package: protoString("storage"),
		MessageType: []*descriptorpb.DescriptorProto{
			fileInfo,
			uploadReq,
		},
	}
	fd, err := protodesc.NewFile(file, nil)
	if err != nil {
		t.Fatalf("build descriptor: %v", err)
	}
	return fd.Messages().ByName("UploadReq")
}

func protoString(v string) *string { return &v }
func protoInt32(v int32) *int32    { return &v }
func protoFieldType(v descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto_Type {
	return &v
}

func TestMethodCtxResponseBody(t *testing.T) {
	md := &methodCtx{}
	md.setResponseBodyField("Data")
	if md.ResponseBody != ".Data" {
		t.Fatalf("ResponseBody = %q, want .Data", md.ResponseBody)
	}
	if !md.HasRawResponse {
		t.Fatal("HasRawResponse should be true")
	}
}

func TestClientNameWithBindings(t *testing.T) {
	tests := []struct {
		name   string
		num    int
		total  int
		expect string
	}{
		{"UploadV0", 0, 1, "UploadV0"},
		{"UploadV0", 0, 2, "UploadV0"},
		{"UploadV0", 1, 2, "UploadV0_1"},
	}
	for _, tt := range tests {
		m := &methodCtx{Name: tt.name, Num: tt.num, BindingsForRPC: tt.total}
		if got := m.ClientName(); got != tt.expect {
			t.Errorf("ClientName() = %q, want %q", got, tt.expect)
		}
	}
}

func TestPluginConfigWarnings(t *testing.T) {
	cfg := newPluginConfig(true, "", "", "full", "error")
	if err := cfg.warn("test warning"); err == nil {
		t.Fatal("expected warning to fail with warnings=error")
	}
	cfg = newPluginConfig(true, "", "", "full", "off")
	if err := cfg.warn("ignored"); err != nil {
		t.Fatalf("expected warning to be ignored, got %v", err)
	}
}

func TestPluginConfigModulePaths(t *testing.T) {
	cfg := newPluginConfig(true, "", "example.com/app", "full", "warn")
	if got := string(cfg.httpxPackage()); got != "example.com/app/transport/httpx" {
		t.Fatalf("httpx path = %q", got)
	}
	if got := string(cfg.pluckPackage()); got != "example.com/app/cmd/protoc-gen-go-http/pluck" {
		t.Fatalf("pluck path = %q", got)
	}
}
