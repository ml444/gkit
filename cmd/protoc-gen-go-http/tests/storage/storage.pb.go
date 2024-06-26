// cmd: protoc --proto_path=. --go-http_out=paths=source_relative:. --go_out=paths=source_relative:. -I=$HOME/github.com/ml444/gctl-templates/protos -I=pluck ./tests/storage/storage.proto

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.25.2
// source: storage/storage.proto

package storage

import (
	_ "github.com/ml444/gkit/cmd/protoc-gen-go-http/pluck"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type FileInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// @desc: 文件名称, 可不填
	FileName string `protobuf:"bytes,1,opt,name=file_name,json=fileName,proto3" json:"file_name,omitempty"`
	// @desc: 文件后缀
	FileSuffix string `protobuf:"bytes,2,opt,name=file_suffix,json=fileSuffix,proto3" json:"file_suffix,omitempty"`
}

func (x *FileInfo) Reset() {
	*x = FileInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_storage_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FileInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FileInfo) ProtoMessage() {}

func (x *FileInfo) ProtoReflect() protoreflect.Message {
	mi := &file_storage_storage_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FileInfo.ProtoReflect.Descriptor instead.
func (*FileInfo) Descriptor() ([]byte, []int) {
	return file_storage_storage_proto_rawDescGZIP(), []int{0}
}

func (x *FileInfo) GetFileName() string {
	if x != nil {
		return x.FileName
	}
	return ""
}

func (x *FileInfo) GetFileSuffix() string {
	if x != nil {
		return x.FileSuffix
	}
	return ""
}

type UploadReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FileInfo *FileInfo `protobuf:"bytes,1,opt,name=file_info,json=fileInfo,proto3" json:"file_info,omitempty"`
	// @desc: 上传的文件内容
	FileData []byte `protobuf:"bytes,2,opt,name=file_data,json=fileData,proto3" json:"file_data,omitempty"`
}

func (x *UploadReq) Reset() {
	*x = UploadReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_storage_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadReq) ProtoMessage() {}

func (x *UploadReq) ProtoReflect() protoreflect.Message {
	mi := &file_storage_storage_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadReq.ProtoReflect.Descriptor instead.
func (*UploadReq) Descriptor() ([]byte, []int) {
	return file_storage_storage_proto_rawDescGZIP(), []int{1}
}

func (x *UploadReq) GetFileInfo() *FileInfo {
	if x != nil {
		return x.FileInfo
	}
	return nil
}

func (x *UploadReq) GetFileData() []byte {
	if x != nil {
		return x.FileData
	}
	return nil
}

type UploadRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Url  string `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	Size uint32 `protobuf:"varint,2,opt,name=size,proto3" json:"size,omitempty"`
}

func (x *UploadRsp) Reset() {
	*x = UploadRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_storage_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadRsp) ProtoMessage() {}

func (x *UploadRsp) ProtoReflect() protoreflect.Message {
	mi := &file_storage_storage_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadRsp.ProtoReflect.Descriptor instead.
func (*UploadRsp) Descriptor() ([]byte, []int) {
	return file_storage_storage_proto_rawDescGZIP(), []int{2}
}

func (x *UploadRsp) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *UploadRsp) GetSize() uint32 {
	if x != nil {
		return x.Size
	}
	return 0
}

type DownloadReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Filename string `protobuf:"bytes,1,opt,name=filename,proto3" json:"filename,omitempty"`
}

func (x *DownloadReq) Reset() {
	*x = DownloadReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_storage_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownloadReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadReq) ProtoMessage() {}

func (x *DownloadReq) ProtoReflect() protoreflect.Message {
	mi := &file_storage_storage_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloadReq.ProtoReflect.Descriptor instead.
func (*DownloadReq) Descriptor() ([]byte, []int) {
	return file_storage_storage_proto_rawDescGZIP(), []int{3}
}

func (x *DownloadReq) GetFilename() string {
	if x != nil {
		return x.Filename
	}
	return ""
}

type DownloadRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Headers map[string]string `protobuf:"bytes,1,rep,name=headers,proto3" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Data    []byte            `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *DownloadRsp) Reset() {
	*x = DownloadRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_storage_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownloadRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadRsp) ProtoMessage() {}

func (x *DownloadRsp) ProtoReflect() protoreflect.Message {
	mi := &file_storage_storage_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloadRsp.ProtoReflect.Descriptor instead.
func (*DownloadRsp) Descriptor() ([]byte, []int) {
	return file_storage_storage_proto_rawDescGZIP(), []int{4}
}

func (x *DownloadRsp) GetHeaders() map[string]string {
	if x != nil {
		return x.Headers
	}
	return nil
}

func (x *DownloadRsp) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_storage_storage_proto protoreflect.FileDescriptor

var file_storage_storage_proto_rawDesc = []byte{
	0x0a, 0x15, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65,
	0x1a, 0x11, 0x70, 0x6c, 0x75, 0x63, 0x6b, 0x2f, 0x70, 0x6c, 0x75, 0x63, 0x6b, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x48, 0x0a, 0x08, 0x46, 0x69, 0x6c, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1b, 0x0a,
	0x09, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x66, 0x69,
	0x6c, 0x65, 0x5f, 0x73, 0x75, 0x66, 0x66, 0x69, 0x78, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0a, 0x66, 0x69, 0x6c, 0x65, 0x53, 0x75, 0x66, 0x66, 0x69, 0x78, 0x22, 0x58, 0x0a, 0x09, 0x55,
	0x70, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x71, 0x12, 0x2e, 0x0a, 0x09, 0x66, 0x69, 0x6c, 0x65,
	0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x73, 0x74,
	0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x08,
	0x66, 0x69, 0x6c, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1b, 0x0a, 0x09, 0x66, 0x69, 0x6c, 0x65,
	0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x66, 0x69, 0x6c,
	0x65, 0x44, 0x61, 0x74, 0x61, 0x22, 0x31, 0x0a, 0x09, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x52,
	0x73, 0x70, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x75, 0x72, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0d, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x22, 0x29, 0x0a, 0x0b, 0x44, 0x6f, 0x77, 0x6e,
	0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x71, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x6e,
	0x61, 0x6d, 0x65, 0x22, 0x9a, 0x01, 0x0a, 0x0b, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64,
	0x52, 0x73, 0x70, 0x12, 0x3b, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x44,
	0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x73, 0x70, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73,
	0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x1a, 0x3a, 0x0a, 0x0c, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01,
	0x32, 0xbc, 0x05, 0x0a, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x12, 0x51, 0x0a, 0x08,
	0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x56, 0x30, 0x12, 0x12, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61,
	0x67, 0x65, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x71, 0x1a, 0x12, 0x2e, 0x73,
	0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x73, 0x70,
	0x22, 0x1d, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x17, 0x22, 0x12, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61,
	0x67, 0x65, 0x2f, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2f, 0x76, 0x30, 0x3a, 0x01, 0x2a, 0x12,
	0x83, 0x01, 0x0a, 0x08, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x56, 0x31, 0x12, 0x12, 0x2e, 0x73,
	0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x71,
	0x1a, 0x12, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61,
	0x64, 0x52, 0x73, 0x70, 0x22, 0x4f, 0xaa, 0x51, 0x27, 0x0a, 0x1a, 0x62, 0x18, 0x61, 0x70, 0x70,
	0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x6f, 0x63, 0x74, 0x65, 0x74, 0x2d, 0x73,
	0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x09, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x69, 0x6e, 0x66, 0x6f,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1f, 0x22, 0x12, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65,
	0x2f, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2f, 0x76, 0x31, 0x3a, 0x09, 0x66, 0x69, 0x6c, 0x65,
	0x5f, 0x64, 0x61, 0x74, 0x61, 0x12, 0x6a, 0x0a, 0x08, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x56,
	0x32, 0x12, 0x12, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x55, 0x70, 0x6c, 0x6f,
	0x61, 0x64, 0x52, 0x65, 0x71, 0x1a, 0x12, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e,
	0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x73, 0x70, 0x22, 0x36, 0xaa, 0x51, 0x16, 0x12, 0x09,
	0x66, 0x69, 0x6c, 0x65, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x1a, 0x09, 0x66, 0x69, 0x6c, 0x65, 0x5f,
	0x64, 0x61, 0x74, 0x61, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x17, 0x22, 0x12, 0x2f, 0x73, 0x74, 0x6f,
	0x72, 0x61, 0x67, 0x65, 0x2f, 0x75, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x2f, 0x76, 0x32, 0x3a, 0x01,
	0x2a, 0x12, 0x59, 0x0a, 0x0a, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x56, 0x30, 0x12,
	0x14, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f,
	0x61, 0x64, 0x52, 0x65, 0x71, 0x1a, 0x14, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e,
	0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x73, 0x70, 0x22, 0x1f, 0x82, 0xd3, 0xe4,
	0x93, 0x02, 0x19, 0x22, 0x14, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2f, 0x64, 0x6f,
	0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x2f, 0x76, 0x30, 0x3a, 0x01, 0x2a, 0x12, 0x6b, 0x0a, 0x0a,
	0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x56, 0x31, 0x12, 0x14, 0x2e, 0x73, 0x74, 0x6f,
	0x72, 0x61, 0x67, 0x65, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x71,
	0x1a, 0x14, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x6c,
	0x6f, 0x61, 0x64, 0x52, 0x73, 0x70, 0x22, 0x31, 0xb2, 0x51, 0x09, 0x12, 0x07, 0x68, 0x65, 0x61,
	0x64, 0x65, 0x72, 0x73, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x1f, 0x22, 0x14, 0x2f, 0x73, 0x74, 0x6f,
	0x72, 0x61, 0x67, 0x65, 0x2f, 0x64, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x2f, 0x76, 0x31,
	0x3a, 0x01, 0x2a, 0x62, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0xa3, 0x01, 0x0a, 0x0a, 0x44, 0x6f,
	0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x56, 0x32, 0x12, 0x14, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61,
	0x67, 0x65, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x65, 0x71, 0x1a, 0x14,
	0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61,
	0x64, 0x52, 0x73, 0x70, 0x22, 0x69, 0xb2, 0x51, 0x47, 0x0a, 0x36, 0x32, 0x13, 0x43, 0x6f, 0x6e,
	0x74, 0x65, 0x6e, 0x74, 0x2d, 0x44, 0x69, 0x73, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e,
	0x92, 0x01, 0x1e, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x76,
	0x6e, 0x64, 0x2e, 0x6f, 0x70, 0x65, 0x6e, 0x78, 0x6d, 0x6c, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74,
	0x73, 0x12, 0x07, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x1a, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x82, 0xd3, 0xe4, 0x93, 0x02, 0x19, 0x22, 0x14, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65,
	0x2f, 0x64, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x2f, 0x76, 0x32, 0x3a, 0x01, 0x2a, 0x42,
	0x12, 0x5a, 0x10, 0x70, 0x6b, 0x67, 0x2f, 0x62, 0x61, 0x73, 0x65, 0x2f, 0x73, 0x74, 0x6f, 0x72,
	0x61, 0x67, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_storage_storage_proto_rawDescOnce sync.Once
	file_storage_storage_proto_rawDescData = file_storage_storage_proto_rawDesc
)

func file_storage_storage_proto_rawDescGZIP() []byte {
	file_storage_storage_proto_rawDescOnce.Do(func() {
		file_storage_storage_proto_rawDescData = protoimpl.X.CompressGZIP(file_storage_storage_proto_rawDescData)
	})
	return file_storage_storage_proto_rawDescData
}

var file_storage_storage_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_storage_storage_proto_goTypes = []interface{}{
	(*FileInfo)(nil),    // 0: storage.FileInfo
	(*UploadReq)(nil),   // 1: storage.UploadReq
	(*UploadRsp)(nil),   // 2: storage.UploadRsp
	(*DownloadReq)(nil), // 3: storage.DownloadReq
	(*DownloadRsp)(nil), // 4: storage.DownloadRsp
	nil,                 // 5: storage.DownloadRsp.HeadersEntry
}
var file_storage_storage_proto_depIdxs = []int32{
	0, // 0: storage.UploadReq.file_info:type_name -> storage.FileInfo
	5, // 1: storage.DownloadRsp.headers:type_name -> storage.DownloadRsp.HeadersEntry
	1, // 2: storage.storage.UploadV0:input_type -> storage.UploadReq
	1, // 3: storage.storage.UploadV1:input_type -> storage.UploadReq
	1, // 4: storage.storage.UploadV2:input_type -> storage.UploadReq
	3, // 5: storage.storage.DownloadV0:input_type -> storage.DownloadReq
	3, // 6: storage.storage.DownloadV1:input_type -> storage.DownloadReq
	3, // 7: storage.storage.DownloadV2:input_type -> storage.DownloadReq
	2, // 8: storage.storage.UploadV0:output_type -> storage.UploadRsp
	2, // 9: storage.storage.UploadV1:output_type -> storage.UploadRsp
	2, // 10: storage.storage.UploadV2:output_type -> storage.UploadRsp
	4, // 11: storage.storage.DownloadV0:output_type -> storage.DownloadRsp
	4, // 12: storage.storage.DownloadV1:output_type -> storage.DownloadRsp
	4, // 13: storage.storage.DownloadV2:output_type -> storage.DownloadRsp
	8, // [8:14] is the sub-list for method output_type
	2, // [2:8] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_storage_storage_proto_init() }
func file_storage_storage_proto_init() {
	if File_storage_storage_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_storage_storage_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FileInfo); i {
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
		file_storage_storage_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadReq); i {
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
		file_storage_storage_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadRsp); i {
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
		file_storage_storage_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownloadReq); i {
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
		file_storage_storage_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownloadRsp); i {
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
			RawDescriptor: file_storage_storage_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_storage_storage_proto_goTypes,
		DependencyIndexes: file_storage_storage_proto_depIdxs,
		MessageInfos:      file_storage_storage_proto_msgTypes,
	}.Build()
	File_storage_storage_proto = out.File
	file_storage_storage_proto_rawDesc = nil
	file_storage_storage_proto_goTypes = nil
	file_storage_storage_proto_depIdxs = nil
}
