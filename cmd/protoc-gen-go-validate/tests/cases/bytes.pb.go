// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.0--rc2
// source: cases/bytes.proto

package cases

import (
	reflect "reflect"
	sync "sync"

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

type BytesNone struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesNone) Reset() {
	*x = BytesNone{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesNone) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesNone) ProtoMessage() {}

func (x *BytesNone) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesNone.ProtoReflect.Descriptor instead.
func (*BytesNone) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{0}
}

func (x *BytesNone) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesConst struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesConst) Reset() {
	*x = BytesConst{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesConst) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesConst) ProtoMessage() {}

func (x *BytesConst) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesConst.ProtoReflect.Descriptor instead.
func (*BytesConst) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{1}
}

func (x *BytesConst) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesIn struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesIn) Reset() {
	*x = BytesIn{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesIn) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesIn) ProtoMessage() {}

func (x *BytesIn) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesIn.ProtoReflect.Descriptor instead.
func (*BytesIn) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{2}
}

func (x *BytesIn) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesNotIn struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesNotIn) Reset() {
	*x = BytesNotIn{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesNotIn) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesNotIn) ProtoMessage() {}

func (x *BytesNotIn) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesNotIn.ProtoReflect.Descriptor instead.
func (*BytesNotIn) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{3}
}

func (x *BytesNotIn) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesLen struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesLen) Reset() {
	*x = BytesLen{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesLen) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesLen) ProtoMessage() {}

func (x *BytesLen) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesLen.ProtoReflect.Descriptor instead.
func (*BytesLen) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{4}
}

func (x *BytesLen) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesMinLen struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesMinLen) Reset() {
	*x = BytesMinLen{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesMinLen) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesMinLen) ProtoMessage() {}

func (x *BytesMinLen) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesMinLen.ProtoReflect.Descriptor instead.
func (*BytesMinLen) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{5}
}

func (x *BytesMinLen) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesMaxLen struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesMaxLen) Reset() {
	*x = BytesMaxLen{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesMaxLen) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesMaxLen) ProtoMessage() {}

func (x *BytesMaxLen) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesMaxLen.ProtoReflect.Descriptor instead.
func (*BytesMaxLen) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{6}
}

func (x *BytesMaxLen) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesMinMaxLen struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesMinMaxLen) Reset() {
	*x = BytesMinMaxLen{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesMinMaxLen) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesMinMaxLen) ProtoMessage() {}

func (x *BytesMinMaxLen) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesMinMaxLen.ProtoReflect.Descriptor instead.
func (*BytesMinMaxLen) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{7}
}

func (x *BytesMinMaxLen) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesEqualMinMaxLen struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesEqualMinMaxLen) Reset() {
	*x = BytesEqualMinMaxLen{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesEqualMinMaxLen) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesEqualMinMaxLen) ProtoMessage() {}

func (x *BytesEqualMinMaxLen) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesEqualMinMaxLen.ProtoReflect.Descriptor instead.
func (*BytesEqualMinMaxLen) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{8}
}

func (x *BytesEqualMinMaxLen) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesPattern struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesPattern) Reset() {
	*x = BytesPattern{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesPattern) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesPattern) ProtoMessage() {}

func (x *BytesPattern) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesPattern.ProtoReflect.Descriptor instead.
func (*BytesPattern) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{9}
}

func (x *BytesPattern) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesPrefix struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesPrefix) Reset() {
	*x = BytesPrefix{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesPrefix) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesPrefix) ProtoMessage() {}

func (x *BytesPrefix) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesPrefix.ProtoReflect.Descriptor instead.
func (*BytesPrefix) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{10}
}

func (x *BytesPrefix) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesContains struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesContains) Reset() {
	*x = BytesContains{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesContains) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesContains) ProtoMessage() {}

func (x *BytesContains) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesContains.ProtoReflect.Descriptor instead.
func (*BytesContains) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{11}
}

func (x *BytesContains) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesSuffix struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesSuffix) Reset() {
	*x = BytesSuffix{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesSuffix) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesSuffix) ProtoMessage() {}

func (x *BytesSuffix) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesSuffix.ProtoReflect.Descriptor instead.
func (*BytesSuffix) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{12}
}

func (x *BytesSuffix) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesIP struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesIP) Reset() {
	*x = BytesIP{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesIP) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesIP) ProtoMessage() {}

func (x *BytesIP) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesIP.ProtoReflect.Descriptor instead.
func (*BytesIP) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{13}
}

func (x *BytesIP) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesIPv4 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesIPv4) Reset() {
	*x = BytesIPv4{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[14]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesIPv4) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesIPv4) ProtoMessage() {}

func (x *BytesIPv4) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[14]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesIPv4.ProtoReflect.Descriptor instead.
func (*BytesIPv4) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{14}
}

func (x *BytesIPv4) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesIPv6 struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesIPv6) Reset() {
	*x = BytesIPv6{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[15]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesIPv6) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesIPv6) ProtoMessage() {}

func (x *BytesIPv6) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[15]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesIPv6.ProtoReflect.Descriptor instead.
func (*BytesIPv6) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{15}
}

func (x *BytesIPv6) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

type BytesIPv6Ignore struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val []byte `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *BytesIPv6Ignore) Reset() {
	*x = BytesIPv6Ignore{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_bytes_proto_msgTypes[16]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BytesIPv6Ignore) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BytesIPv6Ignore) ProtoMessage() {}

func (x *BytesIPv6Ignore) ProtoReflect() protoreflect.Message {
	mi := &file_cases_bytes_proto_msgTypes[16]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BytesIPv6Ignore.ProtoReflect.Descriptor instead.
func (*BytesIPv6Ignore) Descriptor() ([]byte, []int) {
	return file_cases_bytes_proto_rawDescGZIP(), []int{16}
}

func (x *BytesIPv6Ignore) GetVal() []byte {
	if x != nil {
		return x.Val
	}
	return nil
}

var File_cases_bytes_proto protoreflect.FileDescriptor

var file_cases_bytes_proto_rawDesc = []byte{
	0x0a, 0x11, 0x63, 0x61, 0x73, 0x65, 0x73, 0x2f, 0x62, 0x79, 0x74, 0x65, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x63, 0x61, 0x73, 0x65, 0x73,
	0x1a, 0x09, 0x76, 0x2f, 0x76, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x1d, 0x0a, 0x09, 0x42,
	0x79, 0x74, 0x65, 0x73, 0x4e, 0x6f, 0x6e, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x2a, 0x0a, 0x0a, 0x42, 0x79,
	0x74, 0x65, 0x73, 0x43, 0x6f, 0x6e, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x0a, 0xa2, 0x5a, 0x07, 0x7a, 0x05, 0x0a, 0x03, 0x66, 0x6f,
	0x6f, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x2c, 0x0a, 0x07, 0x42, 0x79, 0x74, 0x65, 0x73, 0x49,
	0x6e, 0x12, 0x21, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x0f,
	0xa2, 0x5a, 0x0c, 0x7a, 0x0a, 0x42, 0x03, 0x62, 0x61, 0x72, 0x42, 0x03, 0x62, 0x61, 0x7a, 0x52,
	0x03, 0x76, 0x61, 0x6c, 0x22, 0x31, 0x0a, 0x0a, 0x42, 0x79, 0x74, 0x65, 0x73, 0x4e, 0x6f, 0x74,
	0x49, 0x6e, 0x12, 0x23, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x42,
	0x11, 0xa2, 0x5a, 0x0e, 0x7a, 0x0c, 0x4a, 0x04, 0x66, 0x69, 0x7a, 0x7a, 0x4a, 0x04, 0x62, 0x75,
	0x7a, 0x7a, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x25, 0x0a, 0x08, 0x42, 0x79, 0x74, 0x65, 0x73,
	0x4c, 0x65, 0x6e, 0x12, 0x19, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x42, 0x07, 0xa2, 0x5a, 0x04, 0x7a, 0x02, 0x68, 0x03, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x28,
	0x0a, 0x0b, 0x42, 0x79, 0x74, 0x65, 0x73, 0x4d, 0x69, 0x6e, 0x4c, 0x65, 0x6e, 0x12, 0x19, 0x0a,
	0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x07, 0xa2, 0x5a, 0x04, 0x7a,
	0x02, 0x10, 0x03, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x28, 0x0a, 0x0b, 0x42, 0x79, 0x74, 0x65,
	0x73, 0x4d, 0x61, 0x78, 0x4c, 0x65, 0x6e, 0x12, 0x19, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x42, 0x07, 0xa2, 0x5a, 0x04, 0x7a, 0x02, 0x18, 0x05, 0x52, 0x03, 0x76,
	0x61, 0x6c, 0x22, 0x2d, 0x0a, 0x0e, 0x42, 0x79, 0x74, 0x65, 0x73, 0x4d, 0x69, 0x6e, 0x4d, 0x61,
	0x78, 0x4c, 0x65, 0x6e, 0x12, 0x1b, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x42, 0x09, 0xa2, 0x5a, 0x06, 0x7a, 0x04, 0x10, 0x03, 0x18, 0x05, 0x52, 0x03, 0x76, 0x61,
	0x6c, 0x22, 0x32, 0x0a, 0x13, 0x42, 0x79, 0x74, 0x65, 0x73, 0x45, 0x71, 0x75, 0x61, 0x6c, 0x4d,
	0x69, 0x6e, 0x4d, 0x61, 0x78, 0x4c, 0x65, 0x6e, 0x12, 0x1b, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x09, 0xa2, 0x5a, 0x06, 0x7a, 0x04, 0x10, 0x05, 0x18, 0x05,
	0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x31, 0x0a, 0x0c, 0x42, 0x79, 0x74, 0x65, 0x73, 0x50, 0x61,
	0x74, 0x74, 0x65, 0x72, 0x6e, 0x12, 0x21, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0c, 0x42, 0x0f, 0xa2, 0x5a, 0x0c, 0x7a, 0x0a, 0x22, 0x08, 0x5e, 0x5b, 0x00, 0x2d, 0x7f,
	0x5d, 0x2b, 0x24, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x29, 0x0a, 0x0b, 0x42, 0x79, 0x74, 0x65,
	0x73, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x12, 0x1a, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x42, 0x08, 0xa2, 0x5a, 0x05, 0x7a, 0x03, 0x2a, 0x01, 0x99, 0x52, 0x03,
	0x76, 0x61, 0x6c, 0x22, 0x2d, 0x0a, 0x0d, 0x42, 0x79, 0x74, 0x65, 0x73, 0x43, 0x6f, 0x6e, 0x74,
	0x61, 0x69, 0x6e, 0x73, 0x12, 0x1c, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x42, 0x0a, 0xa2, 0x5a, 0x07, 0x7a, 0x05, 0x3a, 0x03, 0x62, 0x61, 0x72, 0x52, 0x03, 0x76,
	0x61, 0x6c, 0x22, 0x2c, 0x0a, 0x0b, 0x42, 0x79, 0x74, 0x65, 0x73, 0x53, 0x75, 0x66, 0x66, 0x69,
	0x78, 0x12, 0x1d, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x0b,
	0xa2, 0x5a, 0x08, 0x7a, 0x06, 0x32, 0x04, 0x62, 0x75, 0x7a, 0x7a, 0x52, 0x03, 0x76, 0x61, 0x6c,
	0x22, 0x24, 0x0a, 0x07, 0x42, 0x79, 0x74, 0x65, 0x73, 0x49, 0x50, 0x12, 0x19, 0x0a, 0x03, 0x76,
	0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x07, 0xa2, 0x5a, 0x04, 0x7a, 0x02, 0x50,
	0x01, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x26, 0x0a, 0x09, 0x42, 0x79, 0x74, 0x65, 0x73, 0x49,
	0x50, 0x76, 0x34, 0x12, 0x19, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x42, 0x07, 0xa2, 0x5a, 0x04, 0x7a, 0x02, 0x58, 0x01, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x26,
	0x0a, 0x09, 0x42, 0x79, 0x74, 0x65, 0x73, 0x49, 0x50, 0x76, 0x36, 0x12, 0x19, 0x0a, 0x03, 0x76,
	0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x07, 0xa2, 0x5a, 0x04, 0x7a, 0x02, 0x60,
	0x01, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x2e, 0x0a, 0x0f, 0x42, 0x79, 0x74, 0x65, 0x73, 0x49,
	0x50, 0x76, 0x36, 0x49, 0x67, 0x6e, 0x6f, 0x72, 0x65, 0x12, 0x1b, 0x0a, 0x03, 0x76, 0x61, 0x6c,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x09, 0xa2, 0x5a, 0x06, 0x7a, 0x04, 0x70, 0x01, 0x60,
	0x01, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x42, 0x47, 0x5a, 0x45, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x6c, 0x34, 0x34, 0x34, 0x2f, 0x67, 0x6b, 0x69, 0x74, 0x2f,
	0x63, 0x6d, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x67,
	0x6f, 0x2d, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x73,
	0x2f, 0x63, 0x61, 0x73, 0x65, 0x73, 0x2f, 0x67, 0x6f, 0x3b, 0x63, 0x61, 0x73, 0x65, 0x73, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cases_bytes_proto_rawDescOnce sync.Once
	file_cases_bytes_proto_rawDescData = file_cases_bytes_proto_rawDesc
)

func file_cases_bytes_proto_rawDescGZIP() []byte {
	file_cases_bytes_proto_rawDescOnce.Do(func() {
		file_cases_bytes_proto_rawDescData = protoimpl.X.CompressGZIP(file_cases_bytes_proto_rawDescData)
	})
	return file_cases_bytes_proto_rawDescData
}

var file_cases_bytes_proto_msgTypes = make([]protoimpl.MessageInfo, 17)
var file_cases_bytes_proto_goTypes = []interface{}{
	(*BytesNone)(nil),           // 0: tests.cases.BytesNone
	(*BytesConst)(nil),          // 1: tests.cases.BytesConst
	(*BytesIn)(nil),             // 2: tests.cases.BytesIn
	(*BytesNotIn)(nil),          // 3: tests.cases.BytesNotIn
	(*BytesLen)(nil),            // 4: tests.cases.BytesLen
	(*BytesMinLen)(nil),         // 5: tests.cases.BytesMinLen
	(*BytesMaxLen)(nil),         // 6: tests.cases.BytesMaxLen
	(*BytesMinMaxLen)(nil),      // 7: tests.cases.BytesMinMaxLen
	(*BytesEqualMinMaxLen)(nil), // 8: tests.cases.BytesEqualMinMaxLen
	(*BytesPattern)(nil),        // 9: tests.cases.BytesPattern
	(*BytesPrefix)(nil),         // 10: tests.cases.BytesPrefix
	(*BytesContains)(nil),       // 11: tests.cases.BytesContains
	(*BytesSuffix)(nil),         // 12: tests.cases.BytesSuffix
	(*BytesIP)(nil),             // 13: tests.cases.BytesIP
	(*BytesIPv4)(nil),           // 14: tests.cases.BytesIPv4
	(*BytesIPv6)(nil),           // 15: tests.cases.BytesIPv6
	(*BytesIPv6Ignore)(nil),     // 16: tests.cases.BytesIPv6Ignore
}
var file_cases_bytes_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_cases_bytes_proto_init() }
func file_cases_bytes_proto_init() {
	if File_cases_bytes_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cases_bytes_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesNone); i {
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
		file_cases_bytes_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesConst); i {
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
		file_cases_bytes_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesIn); i {
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
		file_cases_bytes_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesNotIn); i {
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
		file_cases_bytes_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesLen); i {
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
		file_cases_bytes_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesMinLen); i {
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
		file_cases_bytes_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesMaxLen); i {
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
		file_cases_bytes_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesMinMaxLen); i {
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
		file_cases_bytes_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesEqualMinMaxLen); i {
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
		file_cases_bytes_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesPattern); i {
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
		file_cases_bytes_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesPrefix); i {
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
		file_cases_bytes_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesContains); i {
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
		file_cases_bytes_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesSuffix); i {
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
		file_cases_bytes_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesIP); i {
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
		file_cases_bytes_proto_msgTypes[14].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesIPv4); i {
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
		file_cases_bytes_proto_msgTypes[15].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesIPv6); i {
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
		file_cases_bytes_proto_msgTypes[16].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BytesIPv6Ignore); i {
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
			RawDescriptor: file_cases_bytes_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   17,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_cases_bytes_proto_goTypes,
		DependencyIndexes: file_cases_bytes_proto_depIdxs,
		MessageInfos:      file_cases_bytes_proto_msgTypes,
	}.Build()
	File_cases_bytes_proto = out.File
	file_cases_bytes_proto_rawDesc = nil
	file_cases_bytes_proto_goTypes = nil
	file_cases_bytes_proto_depIdxs = nil
}
