// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.5
// source: cases/kitchen_sink.proto

package cases

import (
	reflect "reflect"
	sync "sync"

	_ "github.com/ml444/gkit/cmd/protoc-gen-go-validate/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ComplexTestEnum int32

const (
	ComplexTestEnum_ComplexZero ComplexTestEnum = 0
	ComplexTestEnum_ComplexONE  ComplexTestEnum = 1
	ComplexTestEnum_ComplexTWO  ComplexTestEnum = 2
)

// Enum value maps for ComplexTestEnum.
var (
	ComplexTestEnum_name = map[int32]string{
		0: "ComplexZero",
		1: "ComplexONE",
		2: "ComplexTWO",
	}
	ComplexTestEnum_value = map[string]int32{
		"ComplexZero": 0,
		"ComplexONE":  1,
		"ComplexTWO":  2,
	}
)

func (x ComplexTestEnum) Enum() *ComplexTestEnum {
	p := new(ComplexTestEnum)
	*p = x
	return p
}

func (x ComplexTestEnum) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ComplexTestEnum) Descriptor() protoreflect.EnumDescriptor {
	return file_cases_kitchen_sink_proto_enumTypes[0].Descriptor()
}

func (ComplexTestEnum) Type() protoreflect.EnumType {
	return &file_cases_kitchen_sink_proto_enumTypes[0]
}

func (x ComplexTestEnum) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ComplexTestEnum.Descriptor instead.
func (ComplexTestEnum) EnumDescriptor() ([]byte, []int) {
	return file_cases_kitchen_sink_proto_rawDescGZIP(), []int{0}
}

type ComplexTestMsg struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Const      string                   `protobuf:"bytes,1,opt,name=const,proto3" json:"const,omitempty"`
	Nested     *ComplexTestMsg          `protobuf:"bytes,2,opt,name=nested,proto3" json:"nested,omitempty"`
	IntConst   int32                    `protobuf:"varint,3,opt,name=int_const,json=intConst,proto3" json:"int_const,omitempty"`
	BoolConst  bool                     `protobuf:"varint,4,opt,name=bool_const,json=boolConst,proto3" json:"bool_const,omitempty"`
	FloatVal   *wrapperspb.FloatValue   `protobuf:"bytes,5,opt,name=float_val,json=floatVal,proto3" json:"float_val,omitempty"`
	DurVal     *durationpb.Duration     `protobuf:"bytes,6,opt,name=dur_val,json=durVal,proto3" json:"dur_val,omitempty"`
	TsVal      *timestamppb.Timestamp   `protobuf:"bytes,7,opt,name=ts_val,json=tsVal,proto3" json:"ts_val,omitempty"`
	Another    *ComplexTestMsg          `protobuf:"bytes,8,opt,name=another,proto3" json:"another,omitempty"`
	FloatConst float32                  `protobuf:"fixed32,9,opt,name=float_const,json=floatConst,proto3" json:"float_const,omitempty"`
	DoubleIn   float64                  `protobuf:"fixed64,10,opt,name=double_in,json=doubleIn,proto3" json:"double_in,omitempty"`
	EnumConst  ComplexTestEnum          `protobuf:"varint,11,opt,name=enum_const,json=enumConst,proto3,enum=tests.cases.ComplexTestEnum" json:"enum_const,omitempty"`
	AnyVal     *anypb.Any               `protobuf:"bytes,12,opt,name=any_val,json=anyVal,proto3" json:"any_val,omitempty"`
	RepTsVal   []*timestamppb.Timestamp `protobuf:"bytes,13,rep,name=rep_ts_val,json=repTsVal,proto3" json:"rep_ts_val,omitempty"`
	MapVal     map[int32]string         `protobuf:"bytes,14,rep,name=map_val,json=mapVal,proto3" json:"map_val,omitempty" protobuf_key:"zigzag32,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	BytesVal   []byte                   `protobuf:"bytes,15,opt,name=bytes_val,json=bytesVal,proto3" json:"bytes_val,omitempty"`
	// Types that are assignable to O:
	//
	//	*ComplexTestMsg_X
	//	*ComplexTestMsg_Y
	O isComplexTestMsg_O `protobuf_oneof:"o"`
}

func (x *ComplexTestMsg) Reset() {
	*x = ComplexTestMsg{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_kitchen_sink_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ComplexTestMsg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ComplexTestMsg) ProtoMessage() {}

func (x *ComplexTestMsg) ProtoReflect() protoreflect.Message {
	mi := &file_cases_kitchen_sink_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ComplexTestMsg.ProtoReflect.Descriptor instead.
func (*ComplexTestMsg) Descriptor() ([]byte, []int) {
	return file_cases_kitchen_sink_proto_rawDescGZIP(), []int{0}
}

func (x *ComplexTestMsg) GetConst() string {
	if x != nil {
		return x.Const
	}
	return ""
}

func (x *ComplexTestMsg) GetNested() *ComplexTestMsg {
	if x != nil {
		return x.Nested
	}
	return nil
}

func (x *ComplexTestMsg) GetIntConst() int32 {
	if x != nil {
		return x.IntConst
	}
	return 0
}

func (x *ComplexTestMsg) GetBoolConst() bool {
	if x != nil {
		return x.BoolConst
	}
	return false
}

func (x *ComplexTestMsg) GetFloatVal() *wrapperspb.FloatValue {
	if x != nil {
		return x.FloatVal
	}
	return nil
}

func (x *ComplexTestMsg) GetDurVal() *durationpb.Duration {
	if x != nil {
		return x.DurVal
	}
	return nil
}

func (x *ComplexTestMsg) GetTsVal() *timestamppb.Timestamp {
	if x != nil {
		return x.TsVal
	}
	return nil
}

func (x *ComplexTestMsg) GetAnother() *ComplexTestMsg {
	if x != nil {
		return x.Another
	}
	return nil
}

func (x *ComplexTestMsg) GetFloatConst() float32 {
	if x != nil {
		return x.FloatConst
	}
	return 0
}

func (x *ComplexTestMsg) GetDoubleIn() float64 {
	if x != nil {
		return x.DoubleIn
	}
	return 0
}

func (x *ComplexTestMsg) GetEnumConst() ComplexTestEnum {
	if x != nil {
		return x.EnumConst
	}
	return ComplexTestEnum_ComplexZero
}

func (x *ComplexTestMsg) GetAnyVal() *anypb.Any {
	if x != nil {
		return x.AnyVal
	}
	return nil
}

func (x *ComplexTestMsg) GetRepTsVal() []*timestamppb.Timestamp {
	if x != nil {
		return x.RepTsVal
	}
	return nil
}

func (x *ComplexTestMsg) GetMapVal() map[int32]string {
	if x != nil {
		return x.MapVal
	}
	return nil
}

func (x *ComplexTestMsg) GetBytesVal() []byte {
	if x != nil {
		return x.BytesVal
	}
	return nil
}

func (m *ComplexTestMsg) GetO() isComplexTestMsg_O {
	if m != nil {
		return m.O
	}
	return nil
}

func (x *ComplexTestMsg) GetX() string {
	if x, ok := x.GetO().(*ComplexTestMsg_X); ok {
		return x.X
	}
	return ""
}

func (x *ComplexTestMsg) GetY() int32 {
	if x, ok := x.GetO().(*ComplexTestMsg_Y); ok {
		return x.Y
	}
	return 0
}

type isComplexTestMsg_O interface {
	isComplexTestMsg_O()
}

type ComplexTestMsg_X struct {
	X string `protobuf:"bytes,16,opt,name=x,proto3,oneof"`
}

type ComplexTestMsg_Y struct {
	Y int32 `protobuf:"varint,17,opt,name=y,proto3,oneof"`
}

func (*ComplexTestMsg_X) isComplexTestMsg_O() {}

func (*ComplexTestMsg_Y) isComplexTestMsg_O() {}

type KitchenSinkMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val *ComplexTestMsg `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *KitchenSinkMessage) Reset() {
	*x = KitchenSinkMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cases_kitchen_sink_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KitchenSinkMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KitchenSinkMessage) ProtoMessage() {}

func (x *KitchenSinkMessage) ProtoReflect() protoreflect.Message {
	mi := &file_cases_kitchen_sink_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KitchenSinkMessage.ProtoReflect.Descriptor instead.
func (*KitchenSinkMessage) Descriptor() ([]byte, []int) {
	return file_cases_kitchen_sink_proto_rawDescGZIP(), []int{1}
}

func (x *KitchenSinkMessage) GetVal() *ComplexTestMsg {
	if x != nil {
		return x.Val
	}
	return nil
}

var File_cases_kitchen_sink_proto protoreflect.FileDescriptor

var file_cases_kitchen_sink_proto_rawDesc = []byte{
	0x0a, 0x18, 0x63, 0x61, 0x73, 0x65, 0x73, 0x2f, 0x6b, 0x69, 0x74, 0x63, 0x68, 0x65, 0x6e, 0x5f,
	0x73, 0x69, 0x6e, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x74, 0x65, 0x73, 0x74,
	0x73, 0x2e, 0x63, 0x61, 0x73, 0x65, 0x73, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74,
	0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x77, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf3, 0x07, 0x0a,
	0x0e, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x78, 0x54, 0x65, 0x73, 0x74, 0x4d, 0x73, 0x67, 0x12,
	0x21, 0x0a, 0x05, 0x63, 0x6f, 0x6e, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x0b,
	0xfa, 0x42, 0x08, 0x72, 0x06, 0x0a, 0x04, 0x61, 0x62, 0x63, 0x64, 0x52, 0x05, 0x63, 0x6f, 0x6e,
	0x73, 0x74, 0x12, 0x33, 0x0a, 0x06, 0x6e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x63, 0x61, 0x73, 0x65, 0x73,
	0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x78, 0x54, 0x65, 0x73, 0x74, 0x4d, 0x73, 0x67, 0x52,
	0x06, 0x6e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x12, 0x24, 0x0a, 0x09, 0x69, 0x6e, 0x74, 0x5f, 0x63,
	0x6f, 0x6e, 0x73, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x1a,
	0x02, 0x08, 0x05, 0x52, 0x08, 0x69, 0x6e, 0x74, 0x43, 0x6f, 0x6e, 0x73, 0x74, 0x12, 0x26, 0x0a,
	0x0a, 0x62, 0x6f, 0x6f, 0x6c, 0x5f, 0x63, 0x6f, 0x6e, 0x73, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x08, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x6a, 0x02, 0x08, 0x00, 0x52, 0x09, 0x62, 0x6f, 0x6f, 0x6c,
	0x43, 0x6f, 0x6e, 0x73, 0x74, 0x12, 0x44, 0x0a, 0x09, 0x66, 0x6c, 0x6f, 0x61, 0x74, 0x5f, 0x76,
	0x61, 0x6c, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x46, 0x6c, 0x6f, 0x61, 0x74,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x0a, 0xfa, 0x42, 0x07, 0x0a, 0x05, 0x25, 0x00, 0x00, 0x00,
	0x00, 0x52, 0x08, 0x66, 0x6c, 0x6f, 0x61, 0x74, 0x56, 0x61, 0x6c, 0x12, 0x46, 0x0a, 0x07, 0x64,
	0x75, 0x72, 0x5f, 0x76, 0x61, 0x6c, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44,
	0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x12, 0xfa, 0x42, 0x07, 0xaa, 0x01, 0x04, 0x1a,
	0x02, 0x08, 0x11, 0xfa, 0x42, 0x05, 0xaa, 0x01, 0x02, 0x08, 0x01, 0x52, 0x06, 0x64, 0x75, 0x72,
	0x56, 0x61, 0x6c, 0x12, 0x3d, 0x0a, 0x06, 0x74, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42,
	0x0a, 0xfa, 0x42, 0x07, 0xb2, 0x01, 0x04, 0x2a, 0x02, 0x08, 0x07, 0x52, 0x05, 0x74, 0x73, 0x56,
	0x61, 0x6c, 0x12, 0x35, 0x0a, 0x07, 0x61, 0x6e, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x63, 0x61, 0x73, 0x65,
	0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x78, 0x54, 0x65, 0x73, 0x74, 0x4d, 0x73, 0x67,
	0x52, 0x07, 0x61, 0x6e, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x12, 0x2b, 0x0a, 0x0b, 0x66, 0x6c, 0x6f,
	0x61, 0x74, 0x5f, 0x63, 0x6f, 0x6e, 0x73, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x02, 0x42, 0x0a,
	0xfa, 0x42, 0x07, 0x0a, 0x05, 0x15, 0x00, 0x00, 0x00, 0x41, 0x52, 0x0a, 0x66, 0x6c, 0x6f, 0x61,
	0x74, 0x43, 0x6f, 0x6e, 0x73, 0x74, 0x12, 0x34, 0x0a, 0x09, 0x64, 0x6f, 0x75, 0x62, 0x6c, 0x65,
	0x5f, 0x69, 0x6e, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x01, 0x42, 0x17, 0xfa, 0x42, 0x14, 0x12, 0x12,
	0x32, 0x10, 0xb4, 0xc8, 0x76, 0xbe, 0x9f, 0x8c, 0x7c, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc0,
	0x5e, 0x40, 0x52, 0x08, 0x64, 0x6f, 0x75, 0x62, 0x6c, 0x65, 0x49, 0x6e, 0x12, 0x45, 0x0a, 0x0a,
	0x65, 0x6e, 0x75, 0x6d, 0x5f, 0x63, 0x6f, 0x6e, 0x73, 0x74, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x1c, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x63, 0x61, 0x73, 0x65, 0x73, 0x2e, 0x43,
	0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x78, 0x54, 0x65, 0x73, 0x74, 0x45, 0x6e, 0x75, 0x6d, 0x42, 0x08,
	0xfa, 0x42, 0x05, 0x82, 0x01, 0x02, 0x08, 0x02, 0x52, 0x09, 0x65, 0x6e, 0x75, 0x6d, 0x43, 0x6f,
	0x6e, 0x73, 0x74, 0x12, 0x63, 0x0a, 0x07, 0x61, 0x6e, 0x79, 0x5f, 0x76, 0x61, 0x6c, 0x18, 0x0c,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x42, 0x34, 0xfa, 0x42, 0x31, 0xa2,
	0x01, 0x2e, 0x12, 0x2c, 0x74, 0x79, 0x70, 0x65, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x61,
	0x70, 0x69, 0x73, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x06, 0x61, 0x6e, 0x79, 0x56, 0x61, 0x6c, 0x12, 0x4b, 0x0a, 0x0a, 0x72, 0x65, 0x70, 0x5f,
	0x74, 0x73, 0x5f, 0x76, 0x61, 0x6c, 0x18, 0x0d, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42, 0x11, 0xfa, 0x42, 0x0e, 0x92, 0x01, 0x0b,
	0x22, 0x09, 0xb2, 0x01, 0x06, 0x32, 0x04, 0x10, 0xc0, 0x84, 0x3d, 0x52, 0x08, 0x72, 0x65, 0x70,
	0x54, 0x73, 0x56, 0x61, 0x6c, 0x12, 0x4e, 0x0a, 0x07, 0x6d, 0x61, 0x70, 0x5f, 0x76, 0x61, 0x6c,
	0x18, 0x0e, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x27, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x63,
	0x61, 0x73, 0x65, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x78, 0x54, 0x65, 0x73, 0x74,
	0x4d, 0x73, 0x67, 0x2e, 0x4d, 0x61, 0x70, 0x56, 0x61, 0x6c, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x42,
	0x0c, 0xfa, 0x42, 0x09, 0x9a, 0x01, 0x06, 0x22, 0x04, 0x3a, 0x02, 0x10, 0x00, 0x52, 0x06, 0x6d,
	0x61, 0x70, 0x56, 0x61, 0x6c, 0x12, 0x26, 0x0a, 0x09, 0x62, 0x79, 0x74, 0x65, 0x73, 0x5f, 0x76,
	0x61, 0x6c, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x09, 0xfa, 0x42, 0x06, 0x7a, 0x04, 0x0a,
	0x02, 0x00, 0x99, 0x52, 0x08, 0x62, 0x79, 0x74, 0x65, 0x73, 0x56, 0x61, 0x6c, 0x12, 0x0e, 0x0a,
	0x01, 0x78, 0x18, 0x10, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x01, 0x78, 0x12, 0x0e, 0x0a,
	0x01, 0x79, 0x18, 0x11, 0x20, 0x01, 0x28, 0x05, 0x48, 0x00, 0x52, 0x01, 0x79, 0x1a, 0x39, 0x0a,
	0x0b, 0x4d, 0x61, 0x70, 0x56, 0x61, 0x6c, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x11, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x08, 0x0a, 0x01, 0x6f, 0x12, 0x03, 0xf8,
	0x42, 0x01, 0x22, 0x43, 0x0a, 0x12, 0x4b, 0x69, 0x74, 0x63, 0x68, 0x65, 0x6e, 0x53, 0x69, 0x6e,
	0x6b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x2d, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x63, 0x61,
	0x73, 0x65, 0x73, 0x2e, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x78, 0x54, 0x65, 0x73, 0x74, 0x4d,
	0x73, 0x67, 0x52, 0x03, 0x76, 0x61, 0x6c, 0x2a, 0x42, 0x0a, 0x0f, 0x43, 0x6f, 0x6d, 0x70, 0x6c,
	0x65, 0x78, 0x54, 0x65, 0x73, 0x74, 0x45, 0x6e, 0x75, 0x6d, 0x12, 0x0f, 0x0a, 0x0b, 0x43, 0x6f,
	0x6d, 0x70, 0x6c, 0x65, 0x78, 0x5a, 0x65, 0x72, 0x6f, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a, 0x43,
	0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x78, 0x4f, 0x4e, 0x45, 0x10, 0x01, 0x12, 0x0e, 0x0a, 0x0a, 0x43,
	0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x78, 0x54, 0x57, 0x4f, 0x10, 0x02, 0x42, 0x47, 0x5a, 0x45, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x6c, 0x34, 0x34, 0x34, 0x2f,
	0x67, 0x6b, 0x69, 0x74, 0x2f, 0x63, 0x6d, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d,
	0x67, 0x65, 0x6e, 0x2d, 0x67, 0x6f, 0x2d, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f,
	0x74, 0x65, 0x73, 0x74, 0x73, 0x2f, 0x63, 0x61, 0x73, 0x65, 0x73, 0x2f, 0x67, 0x6f, 0x3b, 0x63,
	0x61, 0x73, 0x65, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cases_kitchen_sink_proto_rawDescOnce sync.Once
	file_cases_kitchen_sink_proto_rawDescData = file_cases_kitchen_sink_proto_rawDesc
)

func file_cases_kitchen_sink_proto_rawDescGZIP() []byte {
	file_cases_kitchen_sink_proto_rawDescOnce.Do(func() {
		file_cases_kitchen_sink_proto_rawDescData = protoimpl.X.CompressGZIP(file_cases_kitchen_sink_proto_rawDescData)
	})
	return file_cases_kitchen_sink_proto_rawDescData
}

var file_cases_kitchen_sink_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_cases_kitchen_sink_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_cases_kitchen_sink_proto_goTypes = []interface{}{
	(ComplexTestEnum)(0),          // 0: tests.cases.ComplexTestEnum
	(*ComplexTestMsg)(nil),        // 1: tests.cases.ComplexTestMsg
	(*KitchenSinkMessage)(nil),    // 2: tests.cases.KitchenSinkMessage
	nil,                           // 3: tests.cases.ComplexTestMsg.MapValEntry
	(*wrapperspb.FloatValue)(nil), // 4: google.protobuf.FloatValue
	(*durationpb.Duration)(nil),   // 5: google.protobuf.Duration
	(*timestamppb.Timestamp)(nil), // 6: google.protobuf.Timestamp
	(*anypb.Any)(nil),             // 7: google.protobuf.Any
}
var file_cases_kitchen_sink_proto_depIdxs = []int32{
	1,  // 0: tests.cases.ComplexTestMsg.nested:type_name -> tests.cases.ComplexTestMsg
	4,  // 1: tests.cases.ComplexTestMsg.float_val:type_name -> google.protobuf.FloatValue
	5,  // 2: tests.cases.ComplexTestMsg.dur_val:type_name -> google.protobuf.Duration
	6,  // 3: tests.cases.ComplexTestMsg.ts_val:type_name -> google.protobuf.Timestamp
	1,  // 4: tests.cases.ComplexTestMsg.another:type_name -> tests.cases.ComplexTestMsg
	0,  // 5: tests.cases.ComplexTestMsg.enum_const:type_name -> tests.cases.ComplexTestEnum
	7,  // 6: tests.cases.ComplexTestMsg.any_val:type_name -> google.protobuf.Any
	6,  // 7: tests.cases.ComplexTestMsg.rep_ts_val:type_name -> google.protobuf.Timestamp
	3,  // 8: tests.cases.ComplexTestMsg.map_val:type_name -> tests.cases.ComplexTestMsg.MapValEntry
	1,  // 9: tests.cases.KitchenSinkMessage.val:type_name -> tests.cases.ComplexTestMsg
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_cases_kitchen_sink_proto_init() }
func file_cases_kitchen_sink_proto_init() {
	if File_cases_kitchen_sink_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cases_kitchen_sink_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ComplexTestMsg); i {
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
		file_cases_kitchen_sink_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KitchenSinkMessage); i {
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
	file_cases_kitchen_sink_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*ComplexTestMsg_X)(nil),
		(*ComplexTestMsg_Y)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_cases_kitchen_sink_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_cases_kitchen_sink_proto_goTypes,
		DependencyIndexes: file_cases_kitchen_sink_proto_depIdxs,
		EnumInfos:         file_cases_kitchen_sink_proto_enumTypes,
		MessageInfos:      file_cases_kitchen_sink_proto_msgTypes,
	}.Build()
	File_cases_kitchen_sink_proto = out.File
	file_cases_kitchen_sink_proto_rawDesc = nil
	file_cases_kitchen_sink_proto_goTypes = nil
	file_cases_kitchen_sink_proto_depIdxs = nil
}