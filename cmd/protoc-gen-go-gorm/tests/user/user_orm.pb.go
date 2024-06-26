// Code generated by protoc-gen-go-gorm. DO NOT EDIT.
// versions:
// - protoc-gen-go-gorm v1.0.0
// - protoc             v4.25.2
// source: user/user.proto

package user

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm/clause"
	"gorm.io/hints"
	"time"

	"github.com/ml444/gkit/dbx"
)

const (
	DateTime = "2006-01-02 15:04:05"
	DateOnly = "2006-01-02"
	TimeOnly = "15:04:05"
)

func scanDatetime(dt interface{}, t time.Time, layout string) (d interface{}, err error) {
	switch dt.(type) {
	case string:
		d = t.Format(layout)
	case int32:
		d = int32(t.Unix())
	case int64:
		d = t.Unix()
	case uint32:
		d = uint32(t.Unix())
	case uint64:
		d = uint64(t.Unix())
	case time.Time:
		d = t
	case *timestamppb.Timestamp:
		d = &timestamppb.Timestamp{
			Seconds: t.Unix(),
			Nanos:   int32(t.Nanosecond()),
		}
	case *durationpb.Duration:
		d = &durationpb.Duration{
			Seconds: t.Unix(),
			Nanos:   int32(t.Nanosecond()),
		}
	default:
		err = fmt.Errorf("conversion of [%T] type is not supported", dt)
	}
	return
}

func valueToTime(dt interface{}, layout string) (t time.Time, err error) {
	switch d := dt.(type) {
	case string:
		return time.Parse(layout, d)
	case int32:
		t = time.Unix(int64(d), 0)
	case int64:
		t = time.Unix(d, 0)
	case uint32:
		t = time.Unix(int64(d), 0)
	case uint64:
		t = time.Unix(int64(d), 0)
	case time.Time:
		t = d
	case *timestamppb.Timestamp:
		t = time.Unix(d.Seconds, int64(d.Nanos))
	case *durationpb.Duration:
		t = time.Unix(d.Seconds, int64(d.Nanos))
	default:
		err = fmt.Errorf("conversion of [%T] type is not supported", dt)
	}
	return
}

func jsonMarshal(x interface{}) ([]byte, error) {
	if m, ok := x.(proto.Message); ok {
		return protojson.Marshal(m)
	}
	return json.Marshal(x)
}

func jsonUnmarshal(buf []byte, x interface{}) error {
	if m, ok := x.(proto.Message); ok {
		return protojson.Unmarshal(buf, m)
	}
	return json.Unmarshal(buf, x)
}

func bytesMarshal(x interface{}) ([]byte, error) {
	if m, ok := x.(proto.Message); ok {
		return proto.Marshal(m)
	}
	return json.Marshal(x)
}

func bytesUnmarshal(buf []byte, x interface{}) error {
	if m, ok := x.(proto.Message); ok {
		return proto.Unmarshal(buf, m)
	}
	return json.Unmarshal(buf, x)
}

type TUserInfo struct {
	LoginCount  uint32            `gorm:"type:int;comment:登录次数" json:"login_count"`
	LastLoginIp string            `gorm:"type:varchar(255);comment:最后登录IP" json:"last_login_ip"`
	LastLoginAt uint32            `gorm:"type:timestamp;comment:最后登录时间" json:"last_login_at"`
	GroupIds    UserInfo_GroupIds `gorm:"type:json;comment:分组ID" json:"group_ids"`
}

func (x *TUserInfo) ToSource() dbx.IModel {
	return &UserInfo{
		LoginCount:  x.LoginCount,
		LastLoginIp: x.LastLoginIp,
		LastLoginAt: x.LastLoginAt,
		GroupIds:    []uint64(x.GroupIds),
	}
}

func (x *UserInfo) ToORM() dbx.ITModel {
	return &TUserInfo{
		LoginCount:  x.LoginCount,
		LastLoginIp: x.LastLoginIp,
		LastLoginAt: x.LastLoginAt,
		GroupIds:    UserInfo_GroupIds(x.GroupIds),
	}
}

type UserInfo_GroupIds []uint64

func (x *UserInfo_GroupIds) Scan(src interface{}) error {
	switch buf := src.(type) {
	case []byte:
		return json.Unmarshal(buf, &x)
	default:
		return fmt.Errorf("UserInfo_GroupIds unsupported type [%s] to scan", buf)
	}
}

func (x UserInfo_GroupIds) Value() (driver.Value, error) {
	b, err := json.Marshal(&x)
	return string(b), err
}

func (x *User) TableName() string {
	return "my_user"
}

func (x *TUser) TableName() string {
	return "my_user"
}

type TUser struct {
	Id              uint64               `gorm:"comment:主键;primaryKey" json:"id"`
	IsValidated     bool                 `gorm:"type:bool;comment:是否验证" json:"is_validated"`
	Name            string               `gorm:"type:varchar(255);comment:名称;index:idx_name,class:FULLTEXT;index:idx_age_name,priority:2" json:"name"`
	Age             *uint32              `gorm:"type:int;comment:年龄;index:idx_age_name,priority:1" json:"age"`
	CreatedAt       User_CreatedAt       `gorm:"type:date;comment:创建时间" json:"created_at"`
	UpdatedAt       User_UpdatedAt       `gorm:"type:time;comment:更新时间" json:"updated_at"`
	DeletedAt       User_DeletedAt       `gorm:"type:datetime;comment:删除时间" json:"deleted_at"`
	Detail1         *UserInfo            `gorm:"type:json;comment:详情" json:"detail1"`
	DetailBlob1     *User_DetailBlob     `gorm:"type:blob;comment:详情" json:"detail_blob1"`
	Avatar          User_Avatar          `gorm:"type:blob;comment:头像" json:"avatar"`
	Tags            User_Tags            `gorm:"-:all;comment:标签列表" json:"tags"`
	GroupTags       User_GroupTags       `gorm:"type:json;comment:标签名称" json:"group_tags"`
	ClientLoginInfo User_ClientLoginInfo `gorm:"type:json;comment:多端登陆详情" json:"client_login_info"`
	IgnoreData      User_IgnoreData      `gorm:"-:all;comment:标签" json:"ignore_data"`
	State           User_State           `gorm:"type:int;comment:状态" json:"state"`
	Phone           *string              `gorm:"type:varchar(25);default:;comment:名称;uniqueIndex:uidx_phone" json:"phone"`
}

func (x *TUser) ToSource() dbx.IModel {
	return &User{
		Id:              x.Id,
		IsValidated:     x.IsValidated,
		Name:            x.Name,
		Age:             x.Age,
		CreatedAt:       uint32(x.CreatedAt),
		UpdatedAt:       uint32(x.UpdatedAt),
		DeletedAt:       uint32(x.DeletedAt),
		Detail1:         x.Detail1,
		DetailBlob1:     x.DetailBlob1,
		Avatar:          []byte(x.Avatar),
		Tags:            []string(x.Tags),
		GroupTags:       map[string]uint64(x.GroupTags),
		ClientLoginInfo: map[int32]*UserInfo(x.ClientLoginInfo),
		IgnoreData:      map[uint64]*UserInfo(x.IgnoreData),
		State:           x.State,
		Phone:           x.Phone,
	}
}

func (x *User) ToORM() dbx.ITModel {
	return &TUser{
		Id:              x.Id,
		IsValidated:     x.IsValidated,
		Name:            x.Name,
		Age:             x.Age,
		CreatedAt:       User_CreatedAt(x.CreatedAt),
		UpdatedAt:       User_UpdatedAt(x.UpdatedAt),
		DeletedAt:       User_DeletedAt(x.DeletedAt),
		Detail1:         x.Detail1,
		DetailBlob1:     x.DetailBlob1,
		Avatar:          User_Avatar(x.Avatar),
		Tags:            User_Tags(x.Tags),
		GroupTags:       User_GroupTags(x.GroupTags),
		ClientLoginInfo: User_ClientLoginInfo(x.ClientLoginInfo),
		IgnoreData:      User_IgnoreData(x.IgnoreData),
		State:           x.State,
		Phone:           x.Phone,
	}
}

func (x *TUser) UseIndex2IdxName() clause.Expression {
	return hints.UseIndex("idx_name")
}
func (x *TUser) UseIndexForJoin2IdxAgeName() clause.Expression {
	return hints.UseIndex("idx_age_name").ForJoin()
}
func (x *TUser) ForceIndexForGroupBy2IdxNameIdxAgeName() clause.Expression {
	return hints.ForceIndex("idx_name", "idx_age_name").ForGroupBy()
}
func (x *TUser) IgnoreIndexForOrderBy2IdxNameIdxAgeName() clause.Expression {
	return hints.IgnoreIndex("idx_name", "idx_age_name").ForOrderBy()
}

type User_CreatedAt uint32

func (dt *User_CreatedAt) Scan(value interface{}) (err error) {
	nullTime := &sql.NullTime{}
	err = nullTime.Scan(value)
	if err != nil {
		return err
	}
	var realTyp uint32

	val, err := scanDatetime(realTyp, nullTime.Time, DateOnly)
	if err != nil {
		return fmt.Errorf("scan User_CreatedAt error: %w", err)
	}
	*dt = User_CreatedAt(val.(uint32))
	return
}

func (dt User_CreatedAt) Value() (driver.Value, error) {
	t, err := valueToTime(uint32(dt), DateOnly)
	if err != nil {
		return nil, fmt.Errorf("value User_CreatedAt error: %w", err)
	}
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location()), nil
}

type User_UpdatedAt uint32

func (dt *User_UpdatedAt) Scan(value interface{}) (err error) {
	nullTime := &sql.NullTime{}
	err = nullTime.Scan(value)
	if err != nil {
		return err
	}
	var realTyp uint32
	val, err := scanDatetime(realTyp, nullTime.Time, TimeOnly)
	if err != nil {
		return fmt.Errorf("scan User_UpdatedAt error: %w", err)
	}
	*dt = User_UpdatedAt(val.(uint32))
	return
}

func (dt User_UpdatedAt) Value() (driver.Value, error) {
	t, err := valueToTime(uint32(dt), TimeOnly)
	if err != nil {
		return nil, fmt.Errorf("value User_UpdatedAt error: %w", err)
	}
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location()), nil
}

type User_DeletedAt uint32

func (dt *User_DeletedAt) Scan(value interface{}) (err error) {
	nullTime := &sql.NullTime{}
	err = nullTime.Scan(value)
	if err != nil {
		return err
	}
	var realTyp uint32
	val, err := scanDatetime(realTyp, nullTime.Time, DateTime)
	if err != nil {
		return fmt.Errorf("scan User_DeletedAt error: %w", err)
	}
	*dt = User_DeletedAt(val.(uint32))
	return
}

func (dt User_DeletedAt) Value() (driver.Value, error) {
	t, err := valueToTime(uint32(dt), DateTime)
	if err != nil {
		return nil, fmt.Errorf("value User_DeletedAt error: %w", err)
	}
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location()), nil
}

func (x *UserInfo) Scan(src interface{}) error {
	switch buf := src.(type) {
	case []byte:
		return jsonUnmarshal(buf, x)
	default:
		return fmt.Errorf("UserInfo unsupported type [%s] to scan", buf)
	}
}

func (x UserInfo) Value() (driver.Value, error) {
	b, err := jsonMarshal(&x)
	return string(b), err
}

func (x *User_DetailBlob) Scan(src interface{}) error {
	switch buf := src.(type) {
	case []byte:
		return bytesUnmarshal(buf, x)
	default:
		return fmt.Errorf("User_DetailBlob unsupported type [%s] to scan", buf)
	}
}

func (x User_DetailBlob) Value() (driver.Value, error) {
	b, err := bytesMarshal(&x)
	return b, err
}

type User_Avatar []byte

func (x *User_Avatar) Scan(src interface{}) error {
	switch buf := src.(type) {
	case []byte:
		return bytesUnmarshal(buf, &x)
	default:
		return fmt.Errorf("User_Avatar unsupported type [%s] to scan", buf)
	}
}

func (x User_Avatar) Value() (driver.Value, error) {
	b, err := bytesMarshal(&x)
	return b, err
}

type User_Tags []string

type User_GroupTags map[string]uint64

func (x *User_GroupTags) Scan(src interface{}) error {
	switch buf := src.(type) {
	case []byte:
		return json.Unmarshal(buf, &x)
	default:
		return fmt.Errorf("User_GroupTags unsupported type [%s] to scan", buf)
	}
}

func (x User_GroupTags) Value() (driver.Value, error) {
	b, err := json.Marshal(&x)
	return string(b), err
}

type User_ClientLoginInfo map[int32]*UserInfo

func (x *User_ClientLoginInfo) Scan(src interface{}) error {
	switch buf := src.(type) {
	case []byte:
		return json.Unmarshal(buf, &x)
	default:
		return fmt.Errorf("User_ClientLoginInfo unsupported type [%s] to scan", buf)
	}
}

func (x User_ClientLoginInfo) Value() (driver.Value, error) {
	b, err := json.Marshal(&x)
	return string(b), err
}

type User_IgnoreData map[uint64]*UserInfo

func (x *Record) TableName() string {
	return "my_record"
}

func (x *TRecord) TableName() string {
	return "my_record"
}

type TRecord struct {
	Id        uint64         `gorm:"comment:主键;primaryKey" json:"id"`
	CorpId    uint64         `gorm:"comment:企业ID" json:"corp_id"`
	CreatedAt uint32         `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt uint32         `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt uint32         `gorm:"comment:删除时间" json:"deleted_at"`
	Status1   Record_Status  `gorm:"comment:状态" json:"status1"`
	Content   string         `gorm:"comment:内容" json:"content"`
	Rules     Record_Rules   `gorm:"type:json;comment:规则详情" json:"rules"`
	Status2   *Record_Status `gorm:"comment:状态" json:"status2"`
}

func (x *TRecord) ToSource() dbx.IModel {
	return &Record{
		Id:        x.Id,
		CorpId:    x.CorpId,
		CreatedAt: x.CreatedAt,
		UpdatedAt: x.UpdatedAt,
		DeletedAt: x.DeletedAt,
		Status1:   x.Status1,
		Content:   x.Content,
		Rules:     []*Record_Rule(x.Rules),
		Status2:   x.Status2,
	}
}

func (x *Record) ToORM() dbx.ITModel {
	return &TRecord{
		Id:        x.Id,
		CorpId:    x.CorpId,
		CreatedAt: x.CreatedAt,
		UpdatedAt: x.UpdatedAt,
		DeletedAt: x.DeletedAt,
		Status1:   x.Status1,
		Content:   x.Content,
		Rules:     Record_Rules(x.Rules),
		Status2:   x.Status2,
	}
}

type Record_Rules []*Record_Rule

func (x *Record_Rules) Scan(src interface{}) error {
	switch buf := src.(type) {
	case []byte:
		return json.Unmarshal(buf, &x)
	default:
		return fmt.Errorf("Record_Rules unsupported type [%s] to scan", buf)
	}
}

func (x Record_Rules) Value() (driver.Value, error) {
	b, err := json.Marshal(&x)
	return string(b), err
}
