// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.25.2
// source: cases/wkt_timestamp.proto

package cases

import (
	reflect "reflect"
	sync "sync"

	_ "github.com/ml444/gkit/cmd/protoc-gen-go-validate/v"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type TimestampNone struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampNone) Reset() {
	*x = TimestampNone{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampNone) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampNone) ProtoMessage() {}

func (x *TimestampNone) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampNone.ProtoReflect.Descriptor instead.
func (*TimestampNone) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{0}
}

func (x *TimestampNone) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

type TimestampRequired struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampRequired) Reset() {
	*x = TimestampRequired{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampRequired) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampRequired) ProtoMessage() {}

func (x *TimestampRequired) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampRequired.ProtoReflect.Descriptor instead.
func (*TimestampRequired) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{1}
}

func (x *TimestampRequired) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

type TimestampConst struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampConst) Reset() {
	*x = TimestampConst{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampConst) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampConst) ProtoMessage() {}

func (x *TimestampConst) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampConst.ProtoReflect.Descriptor instead.
func (*TimestampConst) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{2}
}

func (x *TimestampConst) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

type TimestampLT struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampLT) Reset() {
	*x = TimestampLT{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampLT) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampLT) ProtoMessage() {}

func (x *TimestampLT) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampLT.ProtoReflect.Descriptor instead.
func (*TimestampLT) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{3}
}

func (x *TimestampLT) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

type TimestampLTE struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampLTE) Reset() {
	*x = TimestampLTE{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampLTE) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampLTE) ProtoMessage() {}

func (x *TimestampLTE) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampLTE.ProtoReflect.Descriptor instead.
func (*TimestampLTE) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{4}
}

func (x *TimestampLTE) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

type TimestampGT struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampGT) Reset() {
	*x = TimestampGT{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampGT) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampGT) ProtoMessage() {}

func (x *TimestampGT) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampGT.ProtoReflect.Descriptor instead.
func (*TimestampGT) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{5}
}

func (x *TimestampGT) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

type TimestampGTE struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampGTE) Reset() {
	*x = TimestampGTE{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampGTE) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampGTE) ProtoMessage() {}

func (x *TimestampGTE) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampGTE.ProtoReflect.Descriptor instead.
func (*TimestampGTE) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{6}
}

func (x *TimestampGTE) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

type TimestampGTLT struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampGTLT) Reset() {
	*x = TimestampGTLT{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampGTLT) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampGTLT) ProtoMessage() {}

func (x *TimestampGTLT) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampGTLT.ProtoReflect.Descriptor instead.
func (*TimestampGTLT) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{7}
}

func (x *TimestampGTLT) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

type TimestampExLTGT struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampExLTGT) Reset() {
	*x = TimestampExLTGT{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampExLTGT) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampExLTGT) ProtoMessage() {}

func (x *TimestampExLTGT) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampExLTGT.ProtoReflect.Descriptor instead.
func (*TimestampExLTGT) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{8}
}

func (x *TimestampExLTGT) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

type TimestampGTELTE struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampGTELTE) Reset() {
	*x = TimestampGTELTE{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampGTELTE) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampGTELTE) ProtoMessage() {}

func (x *TimestampGTELTE) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampGTELTE.ProtoReflect.Descriptor instead.
func (*TimestampGTELTE) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{9}
}

func (x *TimestampGTELTE) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

type TimestampExGTELTE struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampExGTELTE) Reset() {
	*x = TimestampExGTELTE{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampExGTELTE) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampExGTELTE) ProtoMessage() {}

func (x *TimestampExGTELTE) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampExGTELTE.ProtoReflect.Descriptor instead.
func (*TimestampExGTELTE) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{10}
}

func (x *TimestampExGTELTE) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

type TimestampLTNow struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampLTNow) Reset() {
	*x = TimestampLTNow{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampLTNow) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampLTNow) ProtoMessage() {}

func (x *TimestampLTNow) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampLTNow.ProtoReflect.Descriptor instead.
func (*TimestampLTNow) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{11}
}

func (x *TimestampLTNow) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

type TimestampGTNow struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampGTNow) Reset() {
	*x = TimestampGTNow{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampGTNow) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampGTNow) ProtoMessage() {}

func (x *TimestampGTNow) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampGTNow.ProtoReflect.Descriptor instead.
func (*TimestampGTNow) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{12}
}

func (x *TimestampGTNow) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

type TimestampWithin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampWithin) Reset() {
	*x = TimestampWithin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampWithin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampWithin) ProtoMessage() {}

func (x *TimestampWithin) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampWithin.ProtoReflect.Descriptor instead.
func (*TimestampWithin) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{13}
}

func (x *TimestampWithin) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

type TimestampLTNowWithin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampLTNowWithin) Reset() {
	*x = TimestampLTNowWithin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[14]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampLTNowWithin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampLTNowWithin) ProtoMessage() {}

func (x *TimestampLTNowWithin) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[14]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampLTNowWithin.ProtoReflect.Descriptor instead.
func (*TimestampLTNowWithin) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{14}
}

func (x *TimestampLTNowWithin) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

type TimestampGTNowWithin struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *TimestampGTNowWithin) Reset() {
	*x = TimestampGTNowWithin{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_wkt_timestamp_proto_msgTypes[15]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TimestampGTNowWithin) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TimestampGTNowWithin) ProtoMessage() {}

func (x *TimestampGTNowWithin) ProtoReflect() protoreflect.Message {
	mi := &file_cases_wkt_timestamp_proto_msgTypes[15]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TimestampGTNowWithin.ProtoReflect.Descriptor instead.
func (*TimestampGTNowWithin) Descriptor() ([]byte, []int) {
	return file_cases_wkt_timestamp_proto_rawDescGZIP(), []int{15}
}

func (x *TimestampGTNowWithin) GetVal() *timestamppb.Timestamp {
	if x != nil {
		return x.Val
	}
	return nil
}

var File_cases_wkt_timestamp_proto protoreflect.FileDescriptor

var file_cases_wkt_timestamp_proto_rawDesc = []byte{
	0x0a, 0x19, 0x63, 0x61, 0x73, 0x65, 0x73, 0x2f, 0x77, 0x6b, 0x74, 0x5f, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x74, 0x65, 0x73,
	0x74, 0x73, 0x2e, 0x63, 0x61, 0x73, 0x65, 0x73, 0x1a, 0x09, 0x76, 0x2f, 0x76, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x3d, 0x0a, 0x0d, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x4e, 0x6f, 0x6e, 0x65, 0x12, 0x2c, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x03,
	0x76, 0x61, 0x6c, 0x22, 0x4b, 0x0a, 0x11, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x52, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x12, 0x36, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d,
	0x70, 0x42, 0x08, 0xa2, 0x5a, 0x05, 0xb2, 0x01, 0x02, 0x08, 0x01, 0x52, 0x03, 0x76, 0x61, 0x6c,
	0x22, 0x4a, 0x0a, 0x0e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x43, 0x6f, 0x6e,
	0x73, 0x74, 0x12, 0x38, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x0a, 0xa2, 0x5a, 0x07,
	0xb2, 0x01, 0x04, 0x12, 0x02, 0x08, 0x03, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x45, 0x0a, 0x0b,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x4c, 0x54, 0x12, 0x36, 0x0a, 0x03, 0x76,
	0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x42, 0x08, 0xa2, 0x5a, 0x05, 0xb2, 0x01, 0x02, 0x1a, 0x00, 0x52, 0x03,
	0x76, 0x61, 0x6c, 0x22, 0x48, 0x0a, 0x0c, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x4c, 0x54, 0x45, 0x12, 0x38, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x0a, 0xa2, 0x5a,
	0x07, 0xb2, 0x01, 0x04, 0x22, 0x02, 0x08, 0x01, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x48, 0x0a,
	0x0b, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x47, 0x54, 0x12, 0x39, 0x0a, 0x03,
	0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x0b, 0xa2, 0x5a, 0x08, 0xb2, 0x01, 0x05, 0x2a, 0x03, 0x10,
	0xe8, 0x07, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x4a, 0x0a, 0x0c, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x47, 0x54, 0x45, 0x12, 0x3a, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x42, 0x0c, 0xa2, 0x5a, 0x09, 0xb2, 0x01, 0x06, 0x32, 0x04, 0x10, 0xc0, 0x84, 0x3d, 0x52, 0x03,
	0x76, 0x61, 0x6c, 0x22, 0x4b, 0x0a, 0x0d, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x47, 0x54, 0x4c, 0x54, 0x12, 0x3a, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x0c, 0xa2,
	0x5a, 0x09, 0xb2, 0x01, 0x06, 0x1a, 0x02, 0x08, 0x01, 0x2a, 0x00, 0x52, 0x03, 0x76, 0x61, 0x6c,
	0x22, 0x4d, 0x0a, 0x0f, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x45, 0x78, 0x4c,
	0x54, 0x47, 0x54, 0x12, 0x3a, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x0c, 0xa2, 0x5a,
	0x09, 0xb2, 0x01, 0x06, 0x1a, 0x00, 0x2a, 0x02, 0x08, 0x01, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22,
	0x50, 0x0a, 0x0f, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x47, 0x54, 0x45, 0x4c,
	0x54, 0x45, 0x12, 0x3d, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x0f, 0xa2, 0x5a, 0x0c,
	0xb2, 0x01, 0x09, 0x22, 0x03, 0x08, 0x90, 0x1c, 0x32, 0x02, 0x08, 0x3c, 0x52, 0x03, 0x76, 0x61,
	0x6c, 0x22, 0x52, 0x0a, 0x11, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x45, 0x78,
	0x47, 0x54, 0x45, 0x4c, 0x54, 0x45, 0x12, 0x3d, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42,
	0x0f, 0xa2, 0x5a, 0x0c, 0xb2, 0x01, 0x09, 0x22, 0x02, 0x08, 0x3c, 0x32, 0x03, 0x08, 0x90, 0x1c,
	0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x48, 0x0a, 0x0e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x4c, 0x54, 0x4e, 0x6f, 0x77, 0x12, 0x36, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70,
	0x42, 0x08, 0xa2, 0x5a, 0x05, 0xb2, 0x01, 0x02, 0x38, 0x01, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22,
	0x48, 0x0a, 0x0e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x47, 0x54, 0x4e, 0x6f,
	0x77, 0x12, 0x36, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x08, 0xa2, 0x5a, 0x05, 0xb2,
	0x01, 0x02, 0x40, 0x01, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x4c, 0x0a, 0x0f, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x57, 0x69, 0x74, 0x68, 0x69, 0x6e, 0x12, 0x39, 0x0a, 0x03,
	0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x0b, 0xa2, 0x5a, 0x08, 0xb2, 0x01, 0x05, 0x4a, 0x03, 0x08,
	0x90, 0x1c, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x53, 0x0a, 0x14, 0x54, 0x69, 0x6d, 0x65, 0x73,
	0x74, 0x61, 0x6d, 0x70, 0x4c, 0x54, 0x4e, 0x6f, 0x77, 0x57, 0x69, 0x74, 0x68, 0x69, 0x6e, 0x12,
	0x3b, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x0d, 0xa2, 0x5a, 0x0a, 0xb2, 0x01, 0x07,
	0x38, 0x01, 0x4a, 0x03, 0x08, 0x90, 0x1c, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x22, 0x53, 0x0a, 0x14,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x47, 0x54, 0x4e, 0x6f, 0x77, 0x57, 0x69,
	0x74, 0x68, 0x69, 0x6e, 0x12, 0x3b, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x0d, 0xa2,
	0x5a, 0x0a, 0xb2, 0x01, 0x07, 0x40, 0x01, 0x4a, 0x03, 0x08, 0x90, 0x1c, 0x52, 0x03, 0x76, 0x61,
	0x6c, 0x42, 0x47, 0x5a, 0x45, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x6d, 0x6c, 0x34, 0x34, 0x34, 0x2f, 0x67, 0x6b, 0x69, 0x74, 0x2f, 0x63, 0x6d, 0x64, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x67, 0x6f, 0x2d, 0x76, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2f, 0x63, 0x61, 0x73, 0x65,
	0x73, 0x2f, 0x67, 0x6f, 0x3b, 0x63, 0x61, 0x73, 0x65, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_cases_wkt_timestamp_proto_rawDescOnce sync.Once
	file_cases_wkt_timestamp_proto_rawDescData = file_cases_wkt_timestamp_proto_rawDesc
)

func file_cases_wkt_timestamp_proto_rawDescGZIP() []byte {
	file_cases_wkt_timestamp_proto_rawDescOnce.Do(func() {
		file_cases_wkt_timestamp_proto_rawDescData = protoimpl.X.CompressGZIP(file_cases_wkt_timestamp_proto_rawDescData)
	})
	return file_cases_wkt_timestamp_proto_rawDescData
}

var file_cases_wkt_timestamp_proto_msgTypes = make([]protoimpl.MessageInfo, 16)
var file_cases_wkt_timestamp_proto_goTypes = []interface{}{
	(*TimestampNone)(nil),         // 0: tests.cases.TimestampNone
	(*TimestampRequired)(nil),     // 1: tests.cases.TimestampRequired
	(*TimestampConst)(nil),        // 2: tests.cases.TimestampConst
	(*TimestampLT)(nil),           // 3: tests.cases.TimestampLT
	(*TimestampLTE)(nil),          // 4: tests.cases.TimestampLTE
	(*TimestampGT)(nil),           // 5: tests.cases.TimestampGT
	(*TimestampGTE)(nil),          // 6: tests.cases.TimestampGTE
	(*TimestampGTLT)(nil),         // 7: tests.cases.TimestampGTLT
	(*TimestampExLTGT)(nil),       // 8: tests.cases.TimestampExLTGT
	(*TimestampGTELTE)(nil),       // 9: tests.cases.TimestampGTELTE
	(*TimestampExGTELTE)(nil),     // 10: tests.cases.TimestampExGTELTE
	(*TimestampLTNow)(nil),        // 11: tests.cases.TimestampLTNow
	(*TimestampGTNow)(nil),        // 12: tests.cases.TimestampGTNow
	(*TimestampWithin)(nil),       // 13: tests.cases.TimestampWithin
	(*TimestampLTNowWithin)(nil),  // 14: tests.cases.TimestampLTNowWithin
	(*TimestampGTNowWithin)(nil),  // 15: tests.cases.TimestampGTNowWithin
	(*timestamppb.Timestamp)(nil), // 16: google.protobuf.Timestamp
}
var file_cases_wkt_timestamp_proto_depIdxs = []int32{
	16, // 0: tests.cases.TimestampNone.val:type_name -> google.protobuf.Timestamp
	16, // 1: tests.cases.TimestampRequired.val:type_name -> google.protobuf.Timestamp
	16, // 2: tests.cases.TimestampConst.val:type_name -> google.protobuf.Timestamp
	16, // 3: tests.cases.TimestampLT.val:type_name -> google.protobuf.Timestamp
	16, // 4: tests.cases.TimestampLTE.val:type_name -> google.protobuf.Timestamp
	16, // 5: tests.cases.TimestampGT.val:type_name -> google.protobuf.Timestamp
	16, // 6: tests.cases.TimestampGTE.val:type_name -> google.protobuf.Timestamp
	16, // 7: tests.cases.TimestampGTLT.val:type_name -> google.protobuf.Timestamp
	16, // 8: tests.cases.TimestampExLTGT.val:type_name -> google.protobuf.Timestamp
	16, // 9: tests.cases.TimestampGTELTE.val:type_name -> google.protobuf.Timestamp
	16, // 10: tests.cases.TimestampExGTELTE.val:type_name -> google.protobuf.Timestamp
	16, // 11: tests.cases.TimestampLTNow.val:type_name -> google.protobuf.Timestamp
	16, // 12: tests.cases.TimestampGTNow.val:type_name -> google.protobuf.Timestamp
	16, // 13: tests.cases.TimestampWithin.val:type_name -> google.protobuf.Timestamp
	16, // 14: tests.cases.TimestampLTNowWithin.val:type_name -> google.protobuf.Timestamp
	16, // 15: tests.cases.TimestampGTNowWithin.val:type_name -> google.protobuf.Timestamp
	16, // [16:16] is the sub-list for method output_type
	16, // [16:16] is the sub-list for method input_type
	16, // [16:16] is the sub-list for extension type_name
	16, // [16:16] is the sub-list for extension extendee
	0,  // [0:16] is the sub-list for field type_name
}

func init() { file_cases_wkt_timestamp_proto_init() }
func file_cases_wkt_timestamp_proto_init() {
	if File_cases_wkt_timestamp_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cases_wkt_timestamp_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampNone); i {
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
		file_cases_wkt_timestamp_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampRequired); i {
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
		file_cases_wkt_timestamp_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampConst); i {
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
		file_cases_wkt_timestamp_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampLT); i {
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
		file_cases_wkt_timestamp_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampLTE); i {
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
		file_cases_wkt_timestamp_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampGT); i {
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
		file_cases_wkt_timestamp_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampGTE); i {
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
		file_cases_wkt_timestamp_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampGTLT); i {
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
		file_cases_wkt_timestamp_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampExLTGT); i {
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
		file_cases_wkt_timestamp_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampGTELTE); i {
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
		file_cases_wkt_timestamp_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampExGTELTE); i {
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
		file_cases_wkt_timestamp_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampLTNow); i {
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
		file_cases_wkt_timestamp_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampGTNow); i {
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
		file_cases_wkt_timestamp_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampWithin); i {
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
		file_cases_wkt_timestamp_proto_msgTypes[14].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampLTNowWithin); i {
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
		file_cases_wkt_timestamp_proto_msgTypes[15].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TimestampGTNowWithin); i {
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
			RawDescriptor: file_cases_wkt_timestamp_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   16,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_cases_wkt_timestamp_proto_goTypes,
		DependencyIndexes: file_cases_wkt_timestamp_proto_depIdxs,
		MessageInfos:      file_cases_wkt_timestamp_proto_msgTypes,
	}.Build()
	File_cases_wkt_timestamp_proto = out.File
	file_cases_wkt_timestamp_proto_rawDesc = nil
	file_cases_wkt_timestamp_proto_goTypes = nil
	file_cases_wkt_timestamp_proto_depIdxs = nil
}
