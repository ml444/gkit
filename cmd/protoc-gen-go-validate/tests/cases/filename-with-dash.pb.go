// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.0--rc2
// source: cases/filename-with-dash.proto

package cases

import (
	reflect "reflect"

	_ "github.com/ml444/gkit/cmd/protoc-gen-go-validate/v"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_cases_filename_with_dash_proto protoreflect.FileDescriptor

var file_cases_filename_with_dash_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x63, 0x61, 0x73, 0x65, 0x73, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x6e, 0x61, 0x6d, 0x65,
	0x2d, 0x77, 0x69, 0x74, 0x68, 0x2d, 0x64, 0x61, 0x73, 0x68, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x0b, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x63, 0x61, 0x73, 0x65, 0x73, 0x1a, 0x09, 0x76,
	0x2f, 0x76, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x42, 0x47, 0x5a, 0x45, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x6c, 0x34, 0x34, 0x34, 0x2f, 0x67, 0x6b, 0x69,
	0x74, 0x2f, 0x63, 0x6d, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e,
	0x2d, 0x67, 0x6f, 0x2d, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x74, 0x65, 0x73,
	0x74, 0x73, 0x2f, 0x63, 0x61, 0x73, 0x65, 0x73, 0x2f, 0x67, 0x6f, 0x3b, 0x63, 0x61, 0x73, 0x65,
	0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_cases_filename_with_dash_proto_goTypes = []interface{}{}
var file_cases_filename_with_dash_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_cases_filename_with_dash_proto_init() }
func file_cases_filename_with_dash_proto_init() {
	if File_cases_filename_with_dash_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_cases_filename_with_dash_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_cases_filename_with_dash_proto_goTypes,
		DependencyIndexes: file_cases_filename_with_dash_proto_depIdxs,
	}.Build()
	File_cases_filename_with_dash_proto = out.File
	file_cases_filename_with_dash_proto_rawDesc = nil
	file_cases_filename_with_dash_proto_goTypes = nil
	file_cases_filename_with_dash_proto_depIdxs = nil
}
