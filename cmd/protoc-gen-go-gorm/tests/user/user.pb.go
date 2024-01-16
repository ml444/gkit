//go:generate protoc --go_out=. --go-gorm_out=. -I=$HOME/github.com/ml444/gkit/cmd/protoc-gen-go-gorm -I=$HOME/github.com/ml444/gctl-templates/protos user.proto

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.5
// source: user/user.proto

package user

import (
	_ "github.com/ml444/gkit/cmd/protoc-gen-go-gorm/orm"
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

type User_State int32

const (
	User_StateNil     User_State = 0
	User_StateLogin   User_State = 1 // 登录
	User_StateLogout  User_State = 2 // 登出
	User_StateDeleted User_State = 3 // 删除
)

// Enum value maps for User_State.
var (
	User_State_name = map[int32]string{
		0: "StateNil",
		1: "StateLogin",
		2: "StateLogout",
		3: "StateDeleted",
	}
	User_State_value = map[string]int32{
		"StateNil":     0,
		"StateLogin":   1,
		"StateLogout":  2,
		"StateDeleted": 3,
	}
)

func (x User_State) Enum() *User_State {
	p := new(User_State)
	*p = x
	return p
}

func (x User_State) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (User_State) Descriptor() protoreflect.EnumDescriptor {
	return file_user_user_proto_enumTypes[0].Descriptor()
}

func (User_State) Type() protoreflect.EnumType {
	return &file_user_user_proto_enumTypes[0]
}

func (x User_State) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use User_State.Descriptor instead.
func (User_State) EnumDescriptor() ([]byte, []int) {
	return file_user_user_proto_rawDescGZIP(), []int{1, 0}
}

type Record_Status int32

const (
	Record_StatusNil     Record_Status = 0
	Record_StatusUndo    Record_Status = 1 // 未处理
	Record_StatusDoing   Record_Status = 2 // 处理中
	Record_StatusDone    Record_Status = 3 // 已处理
	Record_StatusIgnored Record_Status = 4 // 已忽略
)

// Enum value maps for Record_Status.
var (
	Record_Status_name = map[int32]string{
		0: "StatusNil",
		1: "StatusUndo",
		2: "StatusDoing",
		3: "StatusDone",
		4: "StatusIgnored",
	}
	Record_Status_value = map[string]int32{
		"StatusNil":     0,
		"StatusUndo":    1,
		"StatusDoing":   2,
		"StatusDone":    3,
		"StatusIgnored": 4,
	}
)

func (x Record_Status) Enum() *Record_Status {
	p := new(Record_Status)
	*p = x
	return p
}

func (x Record_Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Record_Status) Descriptor() protoreflect.EnumDescriptor {
	return file_user_user_proto_enumTypes[1].Descriptor()
}

func (Record_Status) Type() protoreflect.EnumType {
	return &file_user_user_proto_enumTypes[1]
}

func (x Record_Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Record_Status.Descriptor instead.
func (Record_Status) EnumDescriptor() ([]byte, []int) {
	return file_user_user_proto_rawDescGZIP(), []int{2, 0}
}

type UserInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LoginCount  uint32   `protobuf:"varint,1,opt,name=login_count,json=loginCount,proto3" json:"login_count,omitempty"`
	LastLoginIp string   `protobuf:"bytes,2,opt,name=last_login_ip,json=lastLoginIp,proto3" json:"last_login_ip,omitempty"`
	LastLoginAt uint32   `protobuf:"varint,3,opt,name=last_login_at,json=lastLoginAt,proto3" json:"last_login_at,omitempty"`
	GroupIds    []uint64 `protobuf:"varint,4,rep,packed,name=group_ids,json=groupIds,proto3" json:"group_ids,omitempty"`
}

func (x *UserInfo) Reset() {
	*x = UserInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_user_user_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserInfo) ProtoMessage() {}

func (x *UserInfo) ProtoReflect() protoreflect.Message {
	mi := &file_user_user_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserInfo.ProtoReflect.Descriptor instead.
func (*UserInfo) Descriptor() ([]byte, []int) {
	return file_user_user_proto_rawDescGZIP(), []int{0}
}

func (x *UserInfo) GetLoginCount() uint32 {
	if x != nil {
		return x.LoginCount
	}
	return 0
}

func (x *UserInfo) GetLastLoginIp() string {
	if x != nil {
		return x.LastLoginIp
	}
	return ""
}

func (x *UserInfo) GetLastLoginAt() uint32 {
	if x != nil {
		return x.LastLoginAt
	}
	return 0
}

func (x *UserInfo) GetGroupIds() []uint64 {
	if x != nil {
		return x.GroupIds
	}
	return nil
}

type User struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id              uint64               `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	IsValidated     bool                 `protobuf:"varint,2,opt,name=is_validated,json=isValidated,proto3" json:"is_validated,omitempty"`
	Name            string               `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Age             *uint32              `protobuf:"varint,4,opt,name=age,proto3,oneof" json:"age,omitempty"`
	CreatedAt       uint32               `protobuf:"varint,5,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt       uint32               `protobuf:"varint,6,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	DeletedAt       uint32               `protobuf:"varint,7,opt,name=deleted_at,json=deletedAt,proto3" json:"deleted_at,omitempty"`
	Detail1         *UserInfo            `protobuf:"bytes,9,opt,name=detail1,proto3" json:"detail1,omitempty"`
	DetailBlob1     *User_DetailBlob     `protobuf:"bytes,10,opt,name=detail_blob1,json=detailBlob1,proto3" json:"detail_blob1,omitempty"`
	Avatar          []byte               `protobuf:"bytes,11,opt,name=avatar,proto3" json:"avatar,omitempty"`
	Tags            []string             `protobuf:"bytes,12,rep,name=tags,proto3" json:"tags,omitempty"`
	GroupTags       map[string]uint64    `protobuf:"bytes,13,rep,name=group_tags,json=groupTags,proto3" json:"group_tags,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	ClientLoginInfo map[int32]*UserInfo  `protobuf:"bytes,14,rep,name=client_login_info,json=clientLoginInfo,proto3" json:"client_login_info,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	IgnoreData      map[uint64]*UserInfo `protobuf:"bytes,15,rep,name=ignore_data,json=ignoreData,proto3" json:"ignore_data,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	State           User_State           `protobuf:"varint,16,opt,name=state,proto3,enum=user.User_State" json:"state,omitempty"`
	Phone           *string              `protobuf:"bytes,17,opt,name=phone,proto3,oneof" json:"phone,omitempty"`
}

func (x *User) Reset() {
	*x = User{}
	if protoimpl.UnsafeEnabled {
		mi := &file_user_user_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *User) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*User) ProtoMessage() {}

func (x *User) ProtoReflect() protoreflect.Message {
	mi := &file_user_user_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use User.ProtoReflect.Descriptor instead.
func (*User) Descriptor() ([]byte, []int) {
	return file_user_user_proto_rawDescGZIP(), []int{1}
}

func (x *User) GetId() uint64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *User) GetIsValidated() bool {
	if x != nil {
		return x.IsValidated
	}
	return false
}

func (x *User) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *User) GetAge() uint32 {
	if x != nil && x.Age != nil {
		return *x.Age
	}
	return 0
}

func (x *User) GetCreatedAt() uint32 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *User) GetUpdatedAt() uint32 {
	if x != nil {
		return x.UpdatedAt
	}
	return 0
}

func (x *User) GetDeletedAt() uint32 {
	if x != nil {
		return x.DeletedAt
	}
	return 0
}

func (x *User) GetDetail1() *UserInfo {
	if x != nil {
		return x.Detail1
	}
	return nil
}

func (x *User) GetDetailBlob1() *User_DetailBlob {
	if x != nil {
		return x.DetailBlob1
	}
	return nil
}

func (x *User) GetAvatar() []byte {
	if x != nil {
		return x.Avatar
	}
	return nil
}

func (x *User) GetTags() []string {
	if x != nil {
		return x.Tags
	}
	return nil
}

func (x *User) GetGroupTags() map[string]uint64 {
	if x != nil {
		return x.GroupTags
	}
	return nil
}

func (x *User) GetClientLoginInfo() map[int32]*UserInfo {
	if x != nil {
		return x.ClientLoginInfo
	}
	return nil
}

func (x *User) GetIgnoreData() map[uint64]*UserInfo {
	if x != nil {
		return x.IgnoreData
	}
	return nil
}

func (x *User) GetState() User_State {
	if x != nil {
		return x.State
	}
	return User_StateNil
}

func (x *User) GetPhone() string {
	if x != nil && x.Phone != nil {
		return *x.Phone
	}
	return ""
}

type Record struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        uint64         `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	CorpId    uint64         `protobuf:"varint,2,opt,name=corp_id,json=corpId,proto3" json:"corp_id,omitempty"`
	CreatedAt uint32         `protobuf:"varint,3,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt uint32         `protobuf:"varint,4,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	DeletedAt uint32         `protobuf:"varint,5,opt,name=deleted_at,json=deletedAt,proto3" json:"deleted_at,omitempty"`
	Status1   Record_Status  `protobuf:"varint,10,opt,name=status1,proto3,enum=user.Record_Status" json:"status1,omitempty"`
	Content   string         `protobuf:"bytes,11,opt,name=content,proto3" json:"content,omitempty"`
	Rules     []*Record_Rule `protobuf:"bytes,12,rep,name=rules,proto3" json:"rules,omitempty"`
	Status2   *Record_Status `protobuf:"varint,13,opt,name=status2,proto3,enum=user.Record_Status,oneof" json:"status2,omitempty"`
}

func (x *Record) Reset() {
	*x = Record{}
	if protoimpl.UnsafeEnabled {
		mi := &file_user_user_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Record) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Record) ProtoMessage() {}

func (x *Record) ProtoReflect() protoreflect.Message {
	mi := &file_user_user_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Record.ProtoReflect.Descriptor instead.
func (*Record) Descriptor() ([]byte, []int) {
	return file_user_user_proto_rawDescGZIP(), []int{2}
}

func (x *Record) GetId() uint64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Record) GetCorpId() uint64 {
	if x != nil {
		return x.CorpId
	}
	return 0
}

func (x *Record) GetCreatedAt() uint32 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *Record) GetUpdatedAt() uint32 {
	if x != nil {
		return x.UpdatedAt
	}
	return 0
}

func (x *Record) GetDeletedAt() uint32 {
	if x != nil {
		return x.DeletedAt
	}
	return 0
}

func (x *Record) GetStatus1() Record_Status {
	if x != nil {
		return x.Status1
	}
	return Record_StatusNil
}

func (x *Record) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

func (x *Record) GetRules() []*Record_Rule {
	if x != nil {
		return x.Rules
	}
	return nil
}

func (x *Record) GetStatus2() Record_Status {
	if x != nil && x.Status2 != nil {
		return *x.Status2
	}
	return Record_StatusNil
}

type User_DetailBlob struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LoginCount  uint32  `protobuf:"varint,1,opt,name=login_count,json=loginCount,proto3" json:"login_count,omitempty"`
	LastLoginIp string  `protobuf:"bytes,2,opt,name=last_login_ip,json=lastLoginIp,proto3" json:"last_login_ip,omitempty"`
	LastLoginAt uint32  `protobuf:"varint,3,opt,name=last_login_at,json=lastLoginAt,proto3" json:"last_login_at,omitempty"`
	GroupIds    []int64 `protobuf:"zigzag64,4,rep,packed,name=group_ids,json=groupIds,proto3" json:"group_ids,omitempty"`
}

func (x *User_DetailBlob) Reset() {
	*x = User_DetailBlob{}
	if protoimpl.UnsafeEnabled {
		mi := &file_user_user_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *User_DetailBlob) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*User_DetailBlob) ProtoMessage() {}

func (x *User_DetailBlob) ProtoReflect() protoreflect.Message {
	mi := &file_user_user_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use User_DetailBlob.ProtoReflect.Descriptor instead.
func (*User_DetailBlob) Descriptor() ([]byte, []int) {
	return file_user_user_proto_rawDescGZIP(), []int{1, 0}
}

func (x *User_DetailBlob) GetLoginCount() uint32 {
	if x != nil {
		return x.LoginCount
	}
	return 0
}

func (x *User_DetailBlob) GetLastLoginIp() string {
	if x != nil {
		return x.LastLoginIp
	}
	return ""
}

func (x *User_DetailBlob) GetLastLoginAt() uint32 {
	if x != nil {
		return x.LastLoginAt
	}
	return 0
}

func (x *User_DetailBlob) GetGroupIds() []int64 {
	if x != nil {
		return x.GroupIds
	}
	return nil
}

type Record_Rule struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RuleId      uint64 `protobuf:"varint,6,opt,name=rule_id,json=ruleId,proto3" json:"rule_id,omitempty"`
	EquipmentId uint64 `protobuf:"varint,7,opt,name=equipment_id,json=equipmentId,proto3" json:"equipment_id,omitempty"`
	ProjectId   uint64 `protobuf:"varint,8,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
}

func (x *Record_Rule) Reset() {
	*x = Record_Rule{}
	if protoimpl.UnsafeEnabled {
		mi := &file_user_user_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Record_Rule) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Record_Rule) ProtoMessage() {}

func (x *Record_Rule) ProtoReflect() protoreflect.Message {
	mi := &file_user_user_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Record_Rule.ProtoReflect.Descriptor instead.
func (*Record_Rule) Descriptor() ([]byte, []int) {
	return file_user_user_proto_rawDescGZIP(), []int{2, 0}
}

func (x *Record_Rule) GetRuleId() uint64 {
	if x != nil {
		return x.RuleId
	}
	return 0
}

func (x *Record_Rule) GetEquipmentId() uint64 {
	if x != nil {
		return x.EquipmentId
	}
	return 0
}

func (x *Record_Rule) GetProjectId() uint64 {
	if x != nil {
		return x.ProjectId
	}
	return 0
}

var File_user_user_proto protoreflect.FileDescriptor

var file_user_user_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x75, 0x73, 0x65, 0x72, 0x2f, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x04, 0x75, 0x73, 0x65, 0x72, 0x1a, 0x0d, 0x6f, 0x72, 0x6d, 0x2f, 0x6f, 0x72, 0x6d,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x89, 0x02, 0x0a, 0x08, 0x55, 0x73, 0x65, 0x72, 0x49,
	0x6e, 0x66, 0x6f, 0x12, 0x37, 0x0a, 0x0b, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x16, 0x9a, 0x49, 0x13, 0x12, 0x03, 0x69,
	0x6e, 0x74, 0x2a, 0x0c, 0xe7, 0x99, 0xbb, 0xe5, 0xbd, 0x95, 0xe6, 0xac, 0xa1, 0xe6, 0x95, 0xb0,
	0x52, 0x0a, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x45, 0x0a, 0x0d,
	0x6c, 0x61, 0x73, 0x74, 0x5f, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x69, 0x70, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x42, 0x21, 0x9a, 0x49, 0x1e, 0x12, 0x0c, 0x76, 0x61, 0x72, 0x63, 0x68, 0x61,
	0x72, 0x28, 0x32, 0x35, 0x35, 0x29, 0x2a, 0x0e, 0xe6, 0x9c, 0x80, 0xe5, 0x90, 0x8e, 0xe7, 0x99,
	0xbb, 0xe5, 0xbd, 0x95, 0x49, 0x50, 0x52, 0x0b, 0x6c, 0x61, 0x73, 0x74, 0x4c, 0x6f, 0x67, 0x69,
	0x6e, 0x49, 0x70, 0x12, 0x46, 0x0a, 0x0d, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x6c, 0x6f, 0x67, 0x69,
	0x6e, 0x5f, 0x61, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x22, 0x9a, 0x49, 0x1f, 0x12,
	0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2a, 0x12, 0xe6, 0x9c, 0x80, 0xe5,
	0x90, 0x8e, 0xe7, 0x99, 0xbb, 0xe5, 0xbd, 0x95, 0xe6, 0x97, 0xb6, 0xe9, 0x97, 0xb4, 0x52, 0x0b,
	0x6c, 0x61, 0x73, 0x74, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x41, 0x74, 0x12, 0x30, 0x0a, 0x09, 0x67,
	0x72, 0x6f, 0x75, 0x70, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x04, 0x42, 0x13,
	0x9a, 0x49, 0x10, 0x12, 0x04, 0x6a, 0x73, 0x6f, 0x6e, 0x2a, 0x08, 0xe5, 0x88, 0x86, 0xe7, 0xbb,
	0x84, 0x49, 0x44, 0x52, 0x08, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x73, 0x3a, 0x03, 0x98,
	0x49, 0x01, 0x22, 0xcb, 0x0d, 0x0a, 0x04, 0x55, 0x73, 0x65, 0x72, 0x12, 0x1d, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x42, 0x0d, 0x9a, 0x49, 0x0a, 0x2a, 0x06, 0xe4, 0xb8,
	0xbb, 0xe9, 0x94, 0xae, 0x30, 0x01, 0x52, 0x02, 0x69, 0x64, 0x12, 0x3a, 0x0a, 0x0c, 0x69, 0x73,
	0x5f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08,
	0x42, 0x17, 0x9a, 0x49, 0x14, 0x12, 0x04, 0x62, 0x6f, 0x6f, 0x6c, 0x2a, 0x0c, 0xe6, 0x98, 0xaf,
	0xe5, 0x90, 0xa6, 0xe9, 0xaa, 0x8c, 0xe8, 0xaf, 0x81, 0x52, 0x0b, 0x69, 0x73, 0x56, 0x61, 0x6c,
	0x69, 0x64, 0x61, 0x74, 0x65, 0x64, 0x12, 0x5f, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x4b, 0x9a, 0x49, 0x48, 0x12, 0x0c, 0x76, 0x61, 0x72, 0x63, 0x68,
	0x61, 0x72, 0x28, 0x32, 0x35, 0x35, 0x29, 0x2a, 0x06, 0xe5, 0x90, 0x8d, 0xe7, 0xa7, 0xb0, 0x42,
	0x17, 0x69, 0x64, 0x78, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x2c, 0x63, 0x6c, 0x61, 0x73, 0x73, 0x3a,
	0x46, 0x55, 0x4c, 0x4c, 0x54, 0x45, 0x58, 0x54, 0x42, 0x17, 0x69, 0x64, 0x78, 0x5f, 0x61, 0x67,
	0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x2c, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x3a,
	0x32, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x40, 0x0a, 0x03, 0x61, 0x67, 0x65, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0d, 0x42, 0x29, 0x9a, 0x49, 0x26, 0x12, 0x03, 0x69, 0x6e, 0x74, 0x2a, 0x06,
	0xe5, 0xb9, 0xb4, 0xe9, 0xbe, 0x84, 0x42, 0x17, 0x69, 0x64, 0x78, 0x5f, 0x61, 0x67, 0x65, 0x5f,
	0x6e, 0x61, 0x6d, 0x65, 0x2c, 0x70, 0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x3a, 0x31, 0x48,
	0x00, 0x52, 0x03, 0x61, 0x67, 0x65, 0x88, 0x01, 0x01, 0x12, 0x36, 0x0a, 0x0a, 0x63, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x17, 0x9a,
	0x49, 0x14, 0x12, 0x04, 0x64, 0x61, 0x74, 0x65, 0x2a, 0x0c, 0xe5, 0x88, 0x9b, 0xe5, 0xbb, 0xba,
	0xe6, 0x97, 0xb6, 0xe9, 0x97, 0xb4, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41,
	0x74, 0x12, 0x36, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x17, 0x9a, 0x49, 0x14, 0x12, 0x04, 0x74, 0x69, 0x6d, 0x65,
	0x2a, 0x0c, 0xe6, 0x9b, 0xb4, 0xe6, 0x96, 0xb0, 0xe6, 0x97, 0xb6, 0xe9, 0x97, 0xb4, 0x52, 0x09,
	0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x3a, 0x0a, 0x0a, 0x64, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x1b, 0x9a,
	0x49, 0x18, 0x12, 0x08, 0x64, 0x61, 0x74, 0x65, 0x74, 0x69, 0x6d, 0x65, 0x2a, 0x0c, 0xe5, 0x88,
	0xa0, 0xe9, 0x99, 0xa4, 0xe6, 0x97, 0xb6, 0xe9, 0x97, 0xb4, 0x52, 0x09, 0x64, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x3b, 0x0a, 0x07, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x31,
	0x18, 0x09, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x55, 0x73,
	0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x42, 0x11, 0x9a, 0x49, 0x0e, 0x12, 0x04, 0x6a, 0x73, 0x6f,
	0x6e, 0x2a, 0x06, 0xe8, 0xaf, 0xa6, 0xe6, 0x83, 0x85, 0x52, 0x07, 0x64, 0x65, 0x74, 0x61, 0x69,
	0x6c, 0x31, 0x12, 0x4b, 0x0a, 0x0c, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x5f, 0x62, 0x6c, 0x6f,
	0x62, 0x31, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e,
	0x55, 0x73, 0x65, 0x72, 0x2e, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x42, 0x6c, 0x6f, 0x62, 0x42,
	0x11, 0x9a, 0x49, 0x0e, 0x12, 0x04, 0x62, 0x6c, 0x6f, 0x62, 0x2a, 0x06, 0xe8, 0xaf, 0xa6, 0xe6,
	0x83, 0x85, 0x52, 0x0b, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x42, 0x6c, 0x6f, 0x62, 0x31, 0x12,
	0x29, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0c, 0x42,
	0x11, 0x9a, 0x49, 0x0e, 0x12, 0x04, 0x62, 0x6c, 0x6f, 0x62, 0x2a, 0x06, 0xe5, 0xa4, 0xb4, 0xe5,
	0x83, 0x8f, 0x52, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x12, 0x28, 0x0a, 0x04, 0x74, 0x61,
	0x67, 0x73, 0x18, 0x0c, 0x20, 0x03, 0x28, 0x09, 0x42, 0x14, 0x9a, 0x49, 0x11, 0x2a, 0x0c, 0xe6,
	0xa0, 0x87, 0xe7, 0xad, 0xbe, 0xe5, 0x88, 0x97, 0xe8, 0xa1, 0xa8, 0xb8, 0x01, 0x01, 0x52, 0x04,
	0x74, 0x61, 0x67, 0x73, 0x12, 0x51, 0x0a, 0x0a, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x74, 0x61,
	0x67, 0x73, 0x18, 0x0d, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e,
	0x55, 0x73, 0x65, 0x72, 0x2e, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x54, 0x61, 0x67, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x42, 0x17, 0x9a, 0x49, 0x14, 0x12, 0x04, 0x6a, 0x73, 0x6f, 0x6e, 0x2a, 0x0c,
	0xe6, 0xa0, 0x87, 0xe7, 0xad, 0xbe, 0xe5, 0x90, 0x8d, 0xe7, 0xa7, 0xb0, 0x52, 0x09, 0x67, 0x72,
	0x6f, 0x75, 0x70, 0x54, 0x61, 0x67, 0x73, 0x12, 0x6a, 0x0a, 0x11, 0x63, 0x6c, 0x69, 0x65, 0x6e,
	0x74, 0x5f, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x0e, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x2e, 0x43,
	0x6c, 0x69, 0x65, 0x6e, 0x74, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x42, 0x1d, 0x9a, 0x49, 0x1a, 0x12, 0x04, 0x6a, 0x73, 0x6f, 0x6e, 0x2a, 0x12,
	0xe5, 0xa4, 0x9a, 0xe7, 0xab, 0xaf, 0xe7, 0x99, 0xbb, 0xe9, 0x99, 0x86, 0xe8, 0xaf, 0xa6, 0xe6,
	0x83, 0x85, 0x52, 0x0f, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x49,
	0x6e, 0x66, 0x6f, 0x12, 0x4b, 0x0a, 0x0b, 0x69, 0x67, 0x6e, 0x6f, 0x72, 0x65, 0x5f, 0x64, 0x61,
	0x74, 0x61, 0x18, 0x0f, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e,
	0x55, 0x73, 0x65, 0x72, 0x2e, 0x49, 0x67, 0x6e, 0x6f, 0x72, 0x65, 0x44, 0x61, 0x74, 0x61, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x42, 0x0e, 0x9a, 0x49, 0x0b, 0x2a, 0x06, 0xe6, 0xa0, 0x87, 0xe7, 0xad,
	0xbe, 0xb8, 0x01, 0x01, 0x52, 0x0a, 0x69, 0x67, 0x6e, 0x6f, 0x72, 0x65, 0x44, 0x61, 0x74, 0x61,
	0x12, 0x38, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x10, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x10, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x2e, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x42, 0x10, 0x9a, 0x49, 0x0d, 0x12, 0x03, 0x69, 0x6e, 0x74, 0x2a, 0x06, 0xe7, 0x8a, 0xb6,
	0xe6, 0x80, 0x81, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x41, 0x0a, 0x05, 0x70, 0x68,
	0x6f, 0x6e, 0x65, 0x18, 0x11, 0x20, 0x01, 0x28, 0x09, 0x42, 0x26, 0x9a, 0x49, 0x23, 0x12, 0x0b,
	0x76, 0x61, 0x72, 0x63, 0x68, 0x61, 0x72, 0x28, 0x32, 0x35, 0x29, 0x1a, 0x00, 0x2a, 0x06, 0xe5,
	0x90, 0x8d, 0xe7, 0xa7, 0xb0, 0x4a, 0x0a, 0x75, 0x69, 0x64, 0x78, 0x5f, 0x70, 0x68, 0x6f, 0x6e,
	0x65, 0x48, 0x01, 0x52, 0x05, 0x70, 0x68, 0x6f, 0x6e, 0x65, 0x88, 0x01, 0x01, 0x1a, 0x86, 0x02,
	0x0a, 0x0a, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x42, 0x6c, 0x6f, 0x62, 0x12, 0x37, 0x0a, 0x0b,
	0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0d, 0x42, 0x16, 0x9a, 0x49, 0x13, 0x12, 0x03, 0x69, 0x6e, 0x74, 0x2a, 0x0c, 0xe7, 0x99, 0xbb,
	0xe5, 0xbd, 0x95, 0xe6, 0xac, 0xa1, 0xe6, 0x95, 0xb0, 0x52, 0x0a, 0x6c, 0x6f, 0x67, 0x69, 0x6e,
	0x43, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x45, 0x0a, 0x0d, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x6c, 0x6f,
	0x67, 0x69, 0x6e, 0x5f, 0x69, 0x70, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x21, 0x9a, 0x49,
	0x1e, 0x12, 0x0c, 0x76, 0x61, 0x72, 0x63, 0x68, 0x61, 0x72, 0x28, 0x32, 0x35, 0x35, 0x29, 0x2a,
	0x0e, 0xe6, 0x9c, 0x80, 0xe5, 0x90, 0x8e, 0xe7, 0x99, 0xbb, 0xe5, 0xbd, 0x95, 0x49, 0x50, 0x52,
	0x0b, 0x6c, 0x61, 0x73, 0x74, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x49, 0x70, 0x12, 0x46, 0x0a, 0x0d,
	0x6c, 0x61, 0x73, 0x74, 0x5f, 0x6c, 0x6f, 0x67, 0x69, 0x6e, 0x5f, 0x61, 0x74, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0d, 0x42, 0x22, 0x9a, 0x49, 0x1f, 0x12, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2a, 0x12, 0xe6, 0x9c, 0x80, 0xe5, 0x90, 0x8e, 0xe7, 0x99, 0xbb, 0xe5, 0xbd,
	0x95, 0xe6, 0x97, 0xb6, 0xe9, 0x97, 0xb4, 0x52, 0x0b, 0x6c, 0x61, 0x73, 0x74, 0x4c, 0x6f, 0x67,
	0x69, 0x6e, 0x41, 0x74, 0x12, 0x30, 0x0a, 0x09, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x69, 0x64,
	0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x12, 0x42, 0x13, 0x9a, 0x49, 0x10, 0x12, 0x04, 0x6a, 0x73,
	0x6f, 0x6e, 0x2a, 0x08, 0xe5, 0x88, 0x86, 0xe7, 0xbb, 0x84, 0x49, 0x44, 0x52, 0x08, 0x67, 0x72,
	0x6f, 0x75, 0x70, 0x49, 0x64, 0x73, 0x1a, 0x3c, 0x0a, 0x0e, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x54,
	0x61, 0x67, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x1a, 0x52, 0x0a, 0x14, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x4c, 0x6f,
	0x67, 0x69, 0x6e, 0x49, 0x6e, 0x66, 0x6f, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x24,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e,
	0x75, 0x73, 0x65, 0x72, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x4d, 0x0a, 0x0f, 0x49, 0x67, 0x6e, 0x6f,
	0x72, 0x65, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b,
	0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x24, 0x0a,
	0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x75,
	0x73, 0x65, 0x72, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x48, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x74, 0x65,
	0x12, 0x0c, 0x0a, 0x08, 0x53, 0x74, 0x61, 0x74, 0x65, 0x4e, 0x69, 0x6c, 0x10, 0x00, 0x12, 0x0e,
	0x0a, 0x0a, 0x53, 0x74, 0x61, 0x74, 0x65, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x10, 0x01, 0x12, 0x0f,
	0x0a, 0x0b, 0x53, 0x74, 0x61, 0x74, 0x65, 0x4c, 0x6f, 0x67, 0x6f, 0x75, 0x74, 0x10, 0x02, 0x12,
	0x10, 0x0a, 0x0c, 0x53, 0x74, 0x61, 0x74, 0x65, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x10,
	0x03, 0x3a, 0x6b, 0x98, 0x49, 0x01, 0xa2, 0x49, 0x07, 0x6d, 0x79, 0x5f, 0x75, 0x73, 0x65, 0x72,
	0xaa, 0x49, 0x0a, 0x1a, 0x08, 0x69, 0x64, 0x78, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0xaa, 0x49, 0x10,
	0x10, 0x01, 0x1a, 0x0c, 0x69, 0x64, 0x78, 0x5f, 0x61, 0x67, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0xaa, 0x49, 0x1c, 0x08, 0x01, 0x10, 0x02, 0x1a, 0x08, 0x69, 0x64, 0x78, 0x5f, 0x6e, 0x61, 0x6d,
	0x65, 0x1a, 0x0c, 0x69, 0x64, 0x78, 0x5f, 0x61, 0x67, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0xaa,
	0x49, 0x1c, 0x08, 0x02, 0x10, 0x03, 0x1a, 0x08, 0x69, 0x64, 0x78, 0x5f, 0x6e, 0x61, 0x6d, 0x65,
	0x1a, 0x0c, 0x69, 0x64, 0x78, 0x5f, 0x61, 0x67, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x42, 0x06,
	0x0a, 0x04, 0x5f, 0x61, 0x67, 0x65, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x70, 0x68, 0x6f, 0x6e, 0x65,
	0x22, 0xd6, 0x05, 0x0a, 0x06, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x12, 0x1d, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x42, 0x0d, 0x9a, 0x49, 0x0a, 0x2a, 0x06, 0xe4, 0xb8,
	0xbb, 0xe9, 0x94, 0xae, 0x30, 0x01, 0x52, 0x02, 0x69, 0x64, 0x12, 0x26, 0x0a, 0x07, 0x63, 0x6f,
	0x72, 0x70, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x42, 0x0d, 0x9a, 0x49, 0x0a,
	0x2a, 0x08, 0xe4, 0xbc, 0x81, 0xe4, 0xb8, 0x9a, 0x49, 0x44, 0x52, 0x06, 0x63, 0x6f, 0x72, 0x70,
	0x49, 0x64, 0x12, 0x30, 0x0a, 0x0a, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x5f, 0x61, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x11, 0x9a, 0x49, 0x0e, 0x2a, 0x0c, 0xe5, 0x88, 0x9b,
	0xe5, 0xbb, 0xba, 0xe6, 0x97, 0xb6, 0xe9, 0x97, 0xb4, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x64, 0x41, 0x74, 0x12, 0x30, 0x0a, 0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x5f,
	0x61, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x11, 0x9a, 0x49, 0x0e, 0x2a, 0x0c, 0xe6,
	0x9b, 0xb4, 0xe6, 0x96, 0xb0, 0xe6, 0x97, 0xb6, 0xe9, 0x97, 0xb4, 0x52, 0x09, 0x75, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x30, 0x0a, 0x0a, 0x64, 0x65, 0x6c, 0x65, 0x74, 0x65,
	0x64, 0x5f, 0x61, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x11, 0x9a, 0x49, 0x0e, 0x2a,
	0x0c, 0xe5, 0x88, 0xa0, 0xe9, 0x99, 0xa4, 0xe6, 0x97, 0xb6, 0xe9, 0x97, 0xb4, 0x52, 0x09, 0x64,
	0x65, 0x6c, 0x65, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x3a, 0x0a, 0x07, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x31, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x13, 0x2e, 0x75, 0x73, 0x65, 0x72,
	0x2e, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x42, 0x0b,
	0x9a, 0x49, 0x08, 0x2a, 0x06, 0xe7, 0x8a, 0xb6, 0xe6, 0x80, 0x81, 0x52, 0x07, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x31, 0x12, 0x25, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18,
	0x0b, 0x20, 0x01, 0x28, 0x09, 0x42, 0x0b, 0x9a, 0x49, 0x08, 0x2a, 0x06, 0xe5, 0x86, 0x85, 0xe5,
	0xae, 0xb9, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x12, 0x40, 0x0a, 0x05, 0x72,
	0x75, 0x6c, 0x65, 0x73, 0x18, 0x0c, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x75, 0x73, 0x65,
	0x72, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x2e, 0x52, 0x75, 0x6c, 0x65, 0x42, 0x17, 0x9a,
	0x49, 0x14, 0x12, 0x04, 0x6a, 0x73, 0x6f, 0x6e, 0x2a, 0x0c, 0xe8, 0xa7, 0x84, 0xe5, 0x88, 0x99,
	0xe8, 0xaf, 0xa6, 0xe6, 0x83, 0x85, 0x52, 0x05, 0x72, 0x75, 0x6c, 0x65, 0x73, 0x12, 0x3f, 0x0a,
	0x07, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x32, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x13,
	0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x2e, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x42, 0x0b, 0x9a, 0x49, 0x08, 0x2a, 0x06, 0xe7, 0x8a, 0xb6, 0xe6, 0x80, 0x81,
	0x48, 0x00, 0x52, 0x07, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x32, 0x88, 0x01, 0x01, 0x1a, 0x8e,
	0x01, 0x0a, 0x04, 0x52, 0x75, 0x6c, 0x65, 0x12, 0x26, 0x0a, 0x07, 0x72, 0x75, 0x6c, 0x65, 0x5f,
	0x69, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x04, 0x42, 0x0d, 0x9a, 0x49, 0x0a, 0x2a, 0x08, 0xe8,
	0xa7, 0x84, 0xe5, 0x88, 0x99, 0x49, 0x44, 0x52, 0x06, 0x72, 0x75, 0x6c, 0x65, 0x49, 0x64, 0x12,
	0x30, 0x0a, 0x0c, 0x65, 0x71, 0x75, 0x69, 0x70, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x69, 0x64, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x04, 0x42, 0x0d, 0x9a, 0x49, 0x0a, 0x2a, 0x08, 0xe8, 0xae, 0xbe, 0xe5,
	0xa4, 0x87, 0x49, 0x44, 0x52, 0x0b, 0x65, 0x71, 0x75, 0x69, 0x70, 0x6d, 0x65, 0x6e, 0x74, 0x49,
	0x64, 0x12, 0x2c, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x18,
	0x08, 0x20, 0x01, 0x28, 0x04, 0x42, 0x0d, 0x9a, 0x49, 0x0a, 0x2a, 0x08, 0xe9, 0xa1, 0xb9, 0xe7,
	0x9b, 0xae, 0x49, 0x44, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x49, 0x64, 0x22,
	0x5b, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0d, 0x0a, 0x09, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x4e, 0x69, 0x6c, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x55, 0x6e, 0x64, 0x6f, 0x10, 0x01, 0x12, 0x0f, 0x0a, 0x0b, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x44, 0x6f, 0x69, 0x6e, 0x67, 0x10, 0x02, 0x12, 0x0e, 0x0a, 0x0a, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x44, 0x6f, 0x6e, 0x65, 0x10, 0x03, 0x12, 0x11, 0x0a, 0x0d, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x49, 0x67, 0x6e, 0x6f, 0x72, 0x65, 0x64, 0x10, 0x04, 0x3a, 0x0f, 0x98, 0x49,
	0x01, 0xa2, 0x49, 0x09, 0x6d, 0x79, 0x5f, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x42, 0x0a, 0x0a,
	0x08, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x32, 0x42, 0x0d, 0x5a, 0x0b, 0x2e, 0x2f, 0x75,
	0x73, 0x65, 0x72, 0x3b, 0x75, 0x73, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_user_user_proto_rawDescOnce sync.Once
	file_user_user_proto_rawDescData = file_user_user_proto_rawDesc
)

func file_user_user_proto_rawDescGZIP() []byte {
	file_user_user_proto_rawDescOnce.Do(func() {
		file_user_user_proto_rawDescData = protoimpl.X.CompressGZIP(file_user_user_proto_rawDescData)
	})
	return file_user_user_proto_rawDescData
}

var file_user_user_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_user_user_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_user_user_proto_goTypes = []interface{}{
	(User_State)(0),         // 0: user.User.State
	(Record_Status)(0),      // 1: user.Record.Status
	(*UserInfo)(nil),        // 2: user.UserInfo
	(*User)(nil),            // 3: user.User
	(*Record)(nil),          // 4: user.Record
	(*User_DetailBlob)(nil), // 5: user.User.DetailBlob
	nil,                     // 6: user.User.GroupTagsEntry
	nil,                     // 7: user.User.ClientLoginInfoEntry
	nil,                     // 8: user.User.IgnoreDataEntry
	(*Record_Rule)(nil),     // 9: user.Record.Rule
}
var file_user_user_proto_depIdxs = []int32{
	2,  // 0: user.User.detail1:type_name -> user.UserInfo
	5,  // 1: user.User.detail_blob1:type_name -> user.User.DetailBlob
	6,  // 2: user.User.group_tags:type_name -> user.User.GroupTagsEntry
	7,  // 3: user.User.client_login_info:type_name -> user.User.ClientLoginInfoEntry
	8,  // 4: user.User.ignore_data:type_name -> user.User.IgnoreDataEntry
	0,  // 5: user.User.state:type_name -> user.User.State
	1,  // 6: user.Record.status1:type_name -> user.Record.Status
	9,  // 7: user.Record.rules:type_name -> user.Record.Rule
	1,  // 8: user.Record.status2:type_name -> user.Record.Status
	2,  // 9: user.User.ClientLoginInfoEntry.value:type_name -> user.UserInfo
	2,  // 10: user.User.IgnoreDataEntry.value:type_name -> user.UserInfo
	11, // [11:11] is the sub-list for method output_type
	11, // [11:11] is the sub-list for method input_type
	11, // [11:11] is the sub-list for extension type_name
	11, // [11:11] is the sub-list for extension extendee
	0,  // [0:11] is the sub-list for field type_name
}

func init() { file_user_user_proto_init() }
func file_user_user_proto_init() {
	if File_user_user_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_user_user_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserInfo); i {
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
		file_user_user_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*User); i {
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
		file_user_user_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Record); i {
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
		file_user_user_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*User_DetailBlob); i {
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
		file_user_user_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Record_Rule); i {
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
	file_user_user_proto_msgTypes[1].OneofWrappers = []interface{}{}
	file_user_user_proto_msgTypes[2].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_user_user_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_user_user_proto_goTypes,
		DependencyIndexes: file_user_user_proto_depIdxs,
		EnumInfos:         file_user_user_proto_enumTypes,
		MessageInfos:      file_user_user_proto_msgTypes,
	}.Build()
	File_user_user_proto = out.File
	file_user_user_proto_rawDesc = nil
	file_user_user_proto_goTypes = nil
	file_user_user_proto_depIdxs = nil
}
