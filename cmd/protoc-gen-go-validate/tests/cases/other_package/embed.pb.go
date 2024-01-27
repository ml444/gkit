// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.5
// source: cases/other_package/embed.proto

package other_package

import (
	reflect "reflect"
	sync "sync"

	_ "github.com/ml444/gkit/cmd/protoc-gen-go-validate/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Embed_Enumerated int32

const (
	Embed_VALUE Embed_Enumerated = 0
)

// Enum value maps for Embed_Enumerated.
var (
	Embed_Enumerated_name = map[int32]string{
		0: "VALUE",
	}
	Embed_Enumerated_value = map[string]int32{
		"VALUE": 0,
	}
)

func (x Embed_Enumerated) Enum() *Embed_Enumerated {
	p := new(Embed_Enumerated)
	*p = x
	return p
}

func (x Embed_Enumerated) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Embed_Enumerated) Descriptor() protoreflect.EnumDescriptor {
	return file_cases_other_package_embed_proto_enumTypes[0].Descriptor()
}

func (Embed_Enumerated) Type() protoreflect.EnumType {
	return &file_cases_other_package_embed_proto_enumTypes[0]
}

func (x Embed_Enumerated) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Embed_Enumerated.Descriptor instead.
func (Embed_Enumerated) EnumDescriptor() ([]byte, []int) {
	return file_cases_other_package_embed_proto_rawDescGZIP(), []int{0, 0}
}

type Embed_FooNumber int32

const (
	Embed_ZERO Embed_FooNumber = 0
	Embed_ONE  Embed_FooNumber = 1
	Embed_TWO  Embed_FooNumber = 2
)

// Enum value maps for Embed_FooNumber.
var (
	Embed_FooNumber_name = map[int32]string{
		0: "ZERO",
		1: "ONE",
		2: "TWO",
	}
	Embed_FooNumber_value = map[string]int32{
		"ZERO": 0,
		"ONE":  1,
		"TWO":  2,
	}
)

func (x Embed_FooNumber) Enum() *Embed_FooNumber {
	p := new(Embed_FooNumber)
	*p = x
	return p
}

func (x Embed_FooNumber) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Embed_FooNumber) Descriptor() protoreflect.EnumDescriptor {
	return file_cases_other_package_embed_proto_enumTypes[1].Descriptor()
}

func (Embed_FooNumber) Type() protoreflect.EnumType {
	return &file_cases_other_package_embed_proto_enumTypes[1]
}

func (x Embed_FooNumber) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Embed_FooNumber.Descriptor instead.
func (Embed_FooNumber) EnumDescriptor() ([]byte, []int) {
	return file_cases_other_package_embed_proto_rawDescGZIP(), []int{0, 1}
}

type Embed_DoubleEmbed_DoubleEnumerated int32

const (
	Embed_DoubleEmbed_VALUE Embed_DoubleEmbed_DoubleEnumerated = 0
)

// Enum value maps for Embed_DoubleEmbed_DoubleEnumerated.
var (
	Embed_DoubleEmbed_DoubleEnumerated_name = map[int32]string{
		0: "VALUE",
	}
	Embed_DoubleEmbed_DoubleEnumerated_value = map[string]int32{
		"VALUE": 0,
	}
)

func (x Embed_DoubleEmbed_DoubleEnumerated) Enum() *Embed_DoubleEmbed_DoubleEnumerated {
	p := new(Embed_DoubleEmbed_DoubleEnumerated)
	*p = x
	return p
}

func (x Embed_DoubleEmbed_DoubleEnumerated) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Embed_DoubleEmbed_DoubleEnumerated) Descriptor() protoreflect.EnumDescriptor {
	return file_cases_other_package_embed_proto_enumTypes[2].Descriptor()
}

func (Embed_DoubleEmbed_DoubleEnumerated) Type() protoreflect.EnumType {
	return &file_cases_other_package_embed_proto_enumTypes[2]
}

func (x Embed_DoubleEmbed_DoubleEnumerated) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Embed_DoubleEmbed_DoubleEnumerated.Descriptor instead.
func (Embed_DoubleEmbed_DoubleEnumerated) EnumDescriptor() ([]byte, []int) {
	return file_cases_other_package_embed_proto_rawDescGZIP(), []int{0, 0, 0}
}

// Validate message embedding across packages.
type Embed struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val int64 `protobuf:"varint,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *Embed) Reset() {
	*x = Embed{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_other_package_embed_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Embed) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Embed) ProtoMessage() {}

func (x *Embed) ProtoReflect() protoreflect.Message {
	mi := &file_cases_other_package_embed_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Embed.ProtoReflect.Descriptor instead.
func (*Embed) Descriptor() ([]byte, []int) {
	return file_cases_other_package_embed_proto_rawDescGZIP(), []int{0}
}

func (x *Embed) GetVal() int64 {
	if x != nil {
		return x.Val
	}
	return 0
}

type Embed_DoubleEmbed struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Embed_DoubleEmbed) Reset() {
	*x = Embed_DoubleEmbed{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_other_package_embed_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Embed_DoubleEmbed) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Embed_DoubleEmbed) ProtoMessage() {}

func (x *Embed_DoubleEmbed) ProtoReflect() protoreflect.Message {
	mi := &file_cases_other_package_embed_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Embed_DoubleEmbed.ProtoReflect.Descriptor instead.
func (*Embed_DoubleEmbed) Descriptor() ([]byte, []int) {
	return file_cases_other_package_embed_proto_rawDescGZIP(), []int{0, 0}
}

var File_cases_other_package_embed_proto protoreflect.FileDescriptor

var file_cases_other_package_embed_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x63, 0x61, 0x73, 0x65, 0x73, 0x2f, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x5f, 0x70, 0x61,
	0x63, 0x6b, 0x61, 0x67, 0x65, 0x2f, 0x65, 0x6d, 0x62, 0x65, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x19, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x63, 0x61, 0x73, 0x65, 0x73, 0x2e, 0x6f,
	0x74, 0x68, 0x65, 0x72, 0x5f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x1a, 0x17, 0x76, 0x61,
	0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x92, 0x01, 0x0a, 0x05, 0x45, 0x6d, 0x62, 0x65, 0x64, 0x12,
	0x19, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x42, 0x07, 0xfa, 0x42,
	0x04, 0x22, 0x02, 0x20, 0x00, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x1a, 0x2c, 0x0a, 0x0b, 0x44, 0x6f,
	0x75, 0x62, 0x6c, 0x65, 0x45, 0x6d, 0x62, 0x65, 0x64, 0x22, 0x1d, 0x0a, 0x10, 0x44, 0x6f, 0x75,
	0x62, 0x6c, 0x65, 0x45, 0x6e, 0x75, 0x6d, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x12, 0x09, 0x0a,
	0x05, 0x56, 0x41, 0x4c, 0x55, 0x45, 0x10, 0x00, 0x22, 0x17, 0x0a, 0x0a, 0x45, 0x6e, 0x75, 0x6d,
	0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x12, 0x09, 0x0a, 0x05, 0x56, 0x41, 0x4c, 0x55, 0x45, 0x10,
	0x00, 0x22, 0x27, 0x0a, 0x09, 0x46, 0x6f, 0x6f, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x08,
	0x0a, 0x04, 0x5a, 0x45, 0x52, 0x4f, 0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x4f, 0x4e, 0x45, 0x10,
	0x01, 0x12, 0x07, 0x0a, 0x03, 0x54, 0x57, 0x4f, 0x10, 0x02, 0x42, 0x5a, 0x5a, 0x58, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x6c, 0x34, 0x34, 0x34, 0x2f, 0x67,
	0x6b, 0x69, 0x74, 0x2f, 0x63, 0x6d, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67,
	0x65, 0x6e, 0x2d, 0x67, 0x6f, 0x2d, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x74,
	0x65, 0x73, 0x74, 0x73, 0x2f, 0x63, 0x61, 0x73, 0x65, 0x73, 0x2f, 0x6f, 0x74, 0x68, 0x65, 0x72,
	0x5f, 0x70, 0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x3b, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x5f, 0x70,
	0x61, 0x63, 0x6b, 0x61, 0x67, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cases_other_package_embed_proto_rawDescOnce sync.Once
	file_cases_other_package_embed_proto_rawDescData = file_cases_other_package_embed_proto_rawDesc
)

func file_cases_other_package_embed_proto_rawDescGZIP() []byte {
	file_cases_other_package_embed_proto_rawDescOnce.Do(func() {
		file_cases_other_package_embed_proto_rawDescData = protoimpl.X.CompressGZIP(file_cases_other_package_embed_proto_rawDescData)
	})
	return file_cases_other_package_embed_proto_rawDescData
}

var file_cases_other_package_embed_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_cases_other_package_embed_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_cases_other_package_embed_proto_goTypes = []interface{}{
	(Embed_Enumerated)(0),                   // 0: tests.cases.other_package.Embed.Enumerated
	(Embed_FooNumber)(0),                    // 1: tests.cases.other_package.Embed.FooNumber
	(Embed_DoubleEmbed_DoubleEnumerated)(0), // 2: tests.cases.other_package.Embed.DoubleEmbed.DoubleEnumerated
	(*Embed)(nil),                           // 3: tests.cases.other_package.Embed
	(*Embed_DoubleEmbed)(nil),               // 4: tests.cases.other_package.Embed.DoubleEmbed
}
var file_cases_other_package_embed_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_cases_other_package_embed_proto_init() }
func file_cases_other_package_embed_proto_init() {
	if File_cases_other_package_embed_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cases_other_package_embed_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Embed); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_cases_other_package_embed_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Embed_DoubleEmbed); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_cases_other_package_embed_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_cases_other_package_embed_proto_goTypes,
		DependencyIndexes: file_cases_other_package_embed_proto_depIdxs,
		EnumInfos:         file_cases_other_package_embed_proto_enumTypes,
		MessageInfos:      file_cases_other_package_embed_proto_msgTypes,
	}.Build()
	File_cases_other_package_embed_proto = out.File
	file_cases_other_package_embed_proto_rawDesc = nil
	file_cases_other_package_embed_proto_goTypes = nil
	file_cases_other_package_embed_proto_depIdxs = nil
}