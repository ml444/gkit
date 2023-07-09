package dbx

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/listoption"
	log "github.com/ml444/glog"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gorm.io/gorm"
)

const (
	ProtoMessageFieldCreatedAt = "created_at"
	ProtoMessageFieldUpdatedAt = "updated_at"
	ProtoMessageFieldDeletedAt = "deleted_at"
)
const DefaultLimit uint32 = 2000
const DefaultOffset uint32 = 0
const MaxLimit uint32 = 100000

var EnableResetDateTime = true

func DisableResetSysDateTimeField() {
	EnableResetDateTime = false
}

type ModelBase interface {
	ProtoReflect() protoreflect.Message
	GetId() uint64
}

type ProtoUpsertedAt interface {
	ProtoReflect() protoreflect.Message
	GetCreatedAt() uint32
	GetUpdatedAt() uint32
	ProtoDeletedAt
	// GetDeletedAt() uint32
	// ProtoReflect() protoreflect.Message
}
type ProtoDeletedAt interface {
	GetDeletedAt() uint32
}

type Scope struct {
	Tx           *gorm.DB
	model        interface{}
	NotFoundErr  error
	RowsAffected int64
	EnableGUID   bool
}

func NewScope(db *gorm.DB, model interface{}) *Scope {
	tx := db.Model(model)
	if _, ok := model.(ProtoDeletedAt); ok {
		tx = tx.Where("deleted_at = 0")
	}
	return &Scope{
		Tx:         tx,
		model:      model,
		EnableGUID: true,
	}
}

func NewScopeOfPure(db *gorm.DB, model interface{}) *Scope {
	return &Scope{
		Tx:         db.Model(model),
		model:      model,
		EnableGUID: true,
	}
}

func (s *Scope) SetNotFoundErr(notFoundErrCode int32) *Scope {
	s.NotFoundErr = errorx.New(notFoundErrCode)
	return s
}
func (s *Scope) DisableGUID() {
	s.EnableGUID = false
}

func (s *Scope) fillResult() {
	s.RowsAffected = s.Tx.RowsAffected
}

// ResetSysDateTimeField To prevent someone from passing in these three fields by mistake, this method is provided to reset
func (s *Scope) ResetSysDateTimeField(v interface{}) {
	if m, ok := v.(ProtoUpsertedAt); ok {
		fields := m.ProtoReflect().Descriptor().Fields()
		m.ProtoReflect().Set(fields.ByName(ProtoMessageFieldCreatedAt), protoreflect.ValueOfUint32(0))
		m.ProtoReflect().Set(fields.ByName(ProtoMessageFieldUpdatedAt), protoreflect.ValueOfUint32(0))
		m.ProtoReflect().Set(fields.ByName(ProtoMessageFieldDeletedAt), protoreflect.ValueOfUint32(0))
	}
}

func (s *Scope) Create(ctx context.Context, v interface{}) error {
	if EnableResetDateTime {
		s.ResetSysDateTimeField(v)
	}
	// if s.EnableGUID {
	// 	m, ok := v.(ModelBase)
	// 	if ok && m.GetId() == 0 {
	// 		// get guid
	// 		rsp, err := guid.GetGuid(ctx, &guid.GetGuidReq{})
	// 		if err != nil {
	// 			log.Errorf("err: %v", err)
	// 			return err
	// 		}
	// 		protoMsg := m.ProtoReflect()
	// 		protoMsg.Set(protoMsg.Descriptor().Fields().ByJSONName("id"), protoreflect.ValueOfUint64(rsp.Id))
	// 	}
	// }

	return s.Tx.Create(v).Error
}

// CreateInBatches Insert data in batches after splitting data according to batchSize
func (s *Scope) CreateInBatches(values interface{}, batchSize int) error {
	defer s.fillResult()
	return s.Tx.CreateInBatches(values, batchSize).Error
}

// Update updates attributes using callbacks. values must be a struct or map.
func (s *Scope) Update(v interface{}, conds ...interface{}) error {
	if EnableResetDateTime {
		s.ResetSysDateTimeField(v)
	}
	if condsLen := len(conds); condsLen == 1 {
		s.Tx.Where(conds[0])
	} else if condsLen > 1 {
		s.Tx.Where(conds[0], conds[1:])
	}
	s.Tx.Updates(v)
	if s.Tx.Error != nil {
		return s.Tx.Error
	}
	s.fillResult()
	if s.Tx.RowsAffected == 0 {
		log.Warnf("model: %v, conds: %v; RowsAffected: 0", v, conds)
	}
	return nil
}

func (s *Scope) Delete(v interface{}, conds ...interface{}) error {
	defer s.fillResult()
	if _, ok := v.(ProtoDeletedAt); ok {
		if condsLen := len(conds); condsLen == 1 {
			s.Tx.Where(conds[0])
		} else if condsLen > 1 {
			s.Tx.Where(conds[0], conds[1:])
		}
		return s.Tx.UpdateColumn(ProtoMessageFieldDeletedAt, time.Now().Unix()).Error
	}
	return s.Tx.Delete(v, conds).Error
}

func (s *Scope) First(dest interface{}, conds ...interface{}) error {
	err := s.Tx.First(dest, conds).Error
	if err == gorm.ErrRecordNotFound {
		if s.NotFoundErr != nil {
			return s.NotFoundErr
		}
		return errorx.CreateError(http.StatusNotFound, errorx.ErrCodeRecordNotFoundSys, err.Error())
	}
	return err
}

func (s *Scope) Find(dest interface{}, conds ...interface{}) error {
	return s.Tx.Find(dest, conds).Error
}

func (s *Scope) Select(fields ...string) *Scope {
	s.Tx.Select(fields)
	return s
}

func (s *Scope) Like(field string, value string) *Scope {
	s.Tx.Where(fmt.Sprintf("%s LIKE ?", field), "%"+value+"%")
	return s
}

func (s *Scope) LikePrefix(field string, value string) *Scope {
	s.Tx.Where(fmt.Sprintf("%s LIKE ?", field), value+"%")
	return s
}

func (s *Scope) Where(query interface{}, args ...interface{}) *Scope {
	s.Tx.Where(query, args)
	return s
}

func (s *Scope) In(field string, values interface{}) *Scope {
	s.Tx.Where(fmt.Sprintf("%s IN ?", field), values)
	return s
}

func (s *Scope) NotIn(field string, values interface{}) *Scope {
	s.Tx.Where(fmt.Sprintf("%s NOT IN ?", field), values)
	return s
}

// Ne :Where("field != ?", arg)
func (s *Scope) Ne(field string, arg interface{}) *Scope {
	s.Tx.Where(fmt.Sprintf("%s = ?", field), arg)
	return s
}
func (s *Scope) Eq(field string, arg interface{}) *Scope {
	s.Tx.Where(fmt.Sprintf("%s = ?", field), arg)
	return s
}
func (s *Scope) Gt(field string, arg interface{}) *Scope {
	s.Tx.Where(fmt.Sprintf("%s > ?", field), arg)
	return s
}
func (s *Scope) Gte(field string, arg interface{}) *Scope {
	s.Tx.Where(fmt.Sprintf("%s >= ?", field), arg)
	return s
}
func (s *Scope) Lt(field string, arg interface{}) *Scope {
	s.Tx.Where(fmt.Sprintf("%s < ?", field), arg)
	return s
}
func (s *Scope) Lte(field string, arg interface{}) *Scope {
	s.Tx.Where(fmt.Sprintf("%s <= ?", field), arg)
	return s
}

func (s *Scope) Between(field string, arg1, arg2 interface{}) *Scope {
	s.Tx.Where(fmt.Sprintf("%s BETWEEN ? AND ?", field), arg1, arg2)
	return s
}
func (s *Scope) NotBetween(field string, arg1, arg2 interface{}) *Scope {
	s.Tx.Where(fmt.Sprintf("%s NOT BETWEEN ? AND ?", field), arg1, arg2)
	return s
}
func (s *Scope) Or(query interface{}, args ...interface{}) *Scope {
	s.Tx.Or(query, args...)
	return s
}

// Order specify order when retrieving records from database
//
//	db.Order("name DESC")
//	db.Order(clause.OrderByColumn{Column: clause.Column{Name: "name"}, Desc: true})
func (s *Scope) Order(value interface{}) {
	s.Tx = s.Tx.Order(value)
}

func (s *Scope) Group(query string) *Scope {
	s.Tx.Group(query)
	return s
}
func (s *Scope) Having(query interface{}, args ...interface{}) *Scope {
	s.Tx.Having(query, args...)
	return s
}
func (s *Scope) Joins(query string, args ...interface{}) *Scope {
	s.Tx.Joins(query, args...)
	return s
}

func (s *Scope) Count(total *int64) error {
	return s.Tx.Count(total).Error
}

func (s *Scope) Offset(offset int) {
	s.Tx = s.Tx.Offset(offset)
}

func (s *Scope) Limit(limit int) {
	s.Tx = s.Tx.Limit(limit)
}
func (s *Scope) PaginateQuery(opt *listoption.ListOption, list interface{}) (*listoption.Paginate, error) {
	var total int64
	if opt != nil && opt.Offset == 0 && !opt.SkipCount {
		err := s.Tx.Count(&total).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	opt = s.SetOffsetAndLimitByListOption(opt)
	err := s.Find(list)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	p := listoption.Paginate{
		Offset: opt.Offset,
		Limit:  opt.Limit,
		Total:  total,
	}
	return &p, nil
}

func (s *Scope) SetOffsetAndLimitByListOption(opt *listoption.ListOption) *listoption.ListOption {
	if opt != nil {
		if opt.Limit == 0 {
			opt.Limit = DefaultLimit
		} else if opt.Limit > MaxLimit {
			opt.Limit = MaxLimit
		}
	} else {
		opt = &listoption.ListOption{}
		opt.Limit = DefaultLimit
		opt.Offset = DefaultOffset
	}
	s.Tx = s.Tx.Limit(int(opt.Limit)).Offset(int(opt.Offset))
	return opt
}

func (s *Scope) Omit(value ...string) *Scope {
	s.Tx.Omit(value...)
	return s
}

func (s *Scope) Unscoped() *Scope {
	s.Tx.Unscoped()
	return s
}

func (s *Scope) Preload(query string, args ...interface{}) *Scope {
	s.Tx.Preload(query, args)
	return s
}

func (s *Scope) Association(value string) *Scope {
	s.Tx.Association(value)
	return s
}
