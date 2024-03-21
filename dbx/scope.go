package dbx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm/clause"
	"gorm.io/gorm/utils"

	"google.golang.org/protobuf/reflect/protoreflect"
	"gorm.io/gorm"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/log"
)

const (
	ProtoMessageFieldCreatedAt = "created_at"
	ProtoMessageFieldUpdatedAt = "updated_at"
	ProtoMessageFieldDeletedAt = "deleted_at"
)
const DefaultLimit uint32 = 2000
const DefaultOffset uint32 = 0
const MaxLimit uint32 = 100000

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

type GenerateIDFunc func(ctx context.Context, cnt int) []uint64

type OrderColumn struct {
	Field   string
	Desc    bool
	Reorder bool
}

type Scope struct {
	*gorm.DB
	model       interface{}
	NotFoundErr error
	//RowsAffected int64
	idFunc GenerateIDFunc

	resetTime         bool
	ignoreNotFoundErr bool
}

func NewScope(db *gorm.DB, model interface{}) *Scope {
	tx := db.Model(model)
	if _, ok := model.(ProtoDeletedAt); ok {
		tx = tx.Where("deleted_at = 0")
	}
	return &Scope{
		DB:    tx,
		model: model,
	}
}

func NewScopeOfPure(db *gorm.DB, model interface{}) *Scope {
	return &Scope{
		DB:    db.Model(model),
		model: model,
	}
}
func (s *Scope) SetGenerateIDFunc(idFunc GenerateIDFunc) *Scope {
	s.idFunc = idFunc
	return s
}
func (s *Scope) SetNotFoundErr(notFoundErrCode int32) *Scope {
	s.NotFoundErr = errorx.NewWithStatus(http.StatusNotFound, notFoundErrCode)
	return s
}

func (s *Scope) IgnoreNotFoundErr() *Scope {
	s.ignoreNotFoundErr = true
	return s
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

func (s *Scope) Create(v interface{}) error {
	if s.resetTime {
		s.ResetSysDateTimeField(v)
	}
	// if s.idFunc != nil {
	// 	m, ok := v.(ModelBase)
	// 	if ok && m.GetId() == 0 {
	// 		// get id
	// 		protoMsg := m.ProtoReflect()
	// 		protoMsg.Set(protoMsg.Descriptor().Fields().ByJSONName("id"), protoreflect.ValueOfUint64(rsp.Id))
	// 	}
	// }

	return s.DB.Create(v).Error
}

// CreateInBatches Insert data in batches after splitting data according to batchSize
func (s *Scope) CreateInBatches(values interface{}, batchSize int) error {
	return s.DB.CreateInBatches(values, batchSize).Error
}

// Update updates attributes using callbacks. values must be a struct or map.
func (s *Scope) Update(v interface{}, conds ...interface{}) error {
	if s.resetTime {
		s.ResetSysDateTimeField(v)
	}
	if len(conds) > 0 {
		s.DB.Where(conds[0], conds[1:])
	}
	s.DB.Updates(v)
	if s.DB.Error != nil {
		return s.DB.Error
	}
	if s.RowsAffected == 0 {
		log.Warnf("model: %v, RowsAffected: 0", v)
	}
	return nil
}

func (s *Scope) Delete(conds ...interface{}) error {
	if _, ok := s.model.(ProtoDeletedAt); ok {
		if len(conds) > 1 {
			s.DB.Where(conds[0], conds[1:])
		}
		return s.DB.UpdateColumn(ProtoMessageFieldDeletedAt, time.Now().Unix()).Error
	}
	return s.DB.Delete(s.model, conds).Error
}

func (s *Scope) First(dest interface{}, conds ...interface{}) error {
	err := s.DB.First(dest, conds).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if s.ignoreNotFoundErr {
			return nil
		}
		if s.NotFoundErr != nil {
			return s.NotFoundErr
		}
		return GetNotFoundErr(err)
	}
	return err
}

func (s *Scope) Exist(conds ...interface{}) (bool, error) {
	if len(conds) > 0 {
		s.Where(conds[0], conds[1:])
	}
	count, err := s.Count()
	if err != nil {
		return false, err
	}
	if count <= 0 {
		return false, nil
	}
	return true, nil
}

func (s *Scope) Find(dest interface{}, conds ...interface{}) error {
	return s.DB.Find(dest, conds).Error
}

func (s *Scope) Select(fields ...string) *Scope {
	s.DB.Select(fields)
	return s
}

func (s *Scope) Like(field string, value string) *Scope {
	s.DB.Where(fmt.Sprintf("%s LIKE ?", field), "%"+value+"%")
	return s
}

func (s *Scope) LikePrefix(field string, value string) *Scope {
	s.DB.Where(fmt.Sprintf("%s LIKE ?", field), value+"%")
	return s
}

func (s *Scope) Where(query interface{}, args ...interface{}) *Scope {
	s.DB.Where(query, args...)
	return s
}

type QueryOpts struct {
	Selects        []string
	Where          map[string]interface{}
	Between        map[string][2]interface{}
	Like           map[string]string
	Or             map[string]interface{}
	OrLike         map[string]string
	OrBetween      map[string][2]interface{}
	isLikePrefix   bool
	isOrLikePrefix bool
	GroupBys       []string
	OrderBys       []OrderColumn
}

func (s *Scope) Query(opts *QueryOpts) *Scope {
	if len(opts.Selects) > 0 {
		s.Select(opts.Selects...)
	}
	if len(opts.Where) > 0 {
		s.Where(opts.Where)
	}
	if len(opts.Between) > 0 {
		for field, value := range opts.Between {
			s.Between(field, value[0], value[1])
		}
	}
	if len(opts.Like) > 0 {
		for field, value := range opts.Like {
			if opts.isLikePrefix {
				s.LikePrefix(field, value)
			} else {
				s.Like(field, value)
			}
		}
	}
	if len(opts.Or) > 0 {
		s.MultiOr(opts.Or)
	}
	if len(opts.OrLike) > 0 {
		s.MultiOrLike(opts.OrLike, opts.isOrLikePrefix)
	}
	if len(opts.OrBetween) > 0 {
		var queryList []string
		var valueList []interface{}
		for field, values := range opts.OrBetween {
			queryList = append(queryList, fmt.Sprintf("(%s BETWEEN ? AND ?)", field))
			valueList = append(valueList, values[0], values[1])
		}
		query := strings.Join(queryList, " OR ")
		s.Where(query, valueList...)
	}
	if len(opts.GroupBys) > 0 {
		s.Groups(opts.GroupBys...)
	}
	if len(opts.OrderBys) > 0 {
		s.Orders(opts.OrderBys...)
	}
	return s
}

func isNonEmptySlice(v interface{}) bool {
	_v := reflect.ValueOf(v)
	return _v.Kind() == reflect.Slice && _v.Len() > 0
}

func (s *Scope) In(field string, values interface{}) *Scope {
	if !isNonEmptySlice(values) {
		return s
	}
	s.DB.Where(fmt.Sprintf("%s IN ?", field), values)
	return s
}

func (s *Scope) NotIn(field string, values interface{}) *Scope {
	if !isNonEmptySlice(values) {
		return s
	}
	s.DB.Where(fmt.Sprintf("%s NOT IN ?", field), values)
	return s
}

// Ne :Where("field != ?", arg)
func (s *Scope) Ne(field string, arg interface{}) *Scope {
	s.DB.Where(fmt.Sprintf("%s != ?", field), arg)
	return s
}
func (s *Scope) Eq(field string, arg interface{}) *Scope {
	s.DB.Where(fmt.Sprintf("%s = ?", field), arg)
	return s
}
func (s *Scope) Gt(field string, arg interface{}) *Scope {
	s.DB.Where(fmt.Sprintf("%s > ?", field), arg)
	return s
}
func (s *Scope) Gte(field string, arg interface{}) *Scope {
	s.DB.Where(fmt.Sprintf("%s >= ?", field), arg)
	return s
}
func (s *Scope) Lt(field string, arg interface{}) *Scope {
	s.DB.Where(fmt.Sprintf("%s < ?", field), arg)
	return s
}
func (s *Scope) Lte(field string, arg interface{}) *Scope {
	s.DB.Where(fmt.Sprintf("%s <= ?", field), arg)
	return s
}

func (s *Scope) Between(field string, arg1, arg2 interface{}) *Scope {
	s.DB.Where(fmt.Sprintf("%s BETWEEN ? AND ?", field), arg1, arg2)
	return s
}
func (s *Scope) NotBetween(field string, arg1, arg2 interface{}) *Scope {
	s.DB.Where(fmt.Sprintf("%s NOT BETWEEN ? AND ?", field), arg1, arg2)
	return s
}
func (s *Scope) Or(query interface{}, args ...interface{}) *Scope {
	s.DB.Or(query, args...)
	return s
}
func (s *Scope) MultiOr(opts map[string]interface{}) *Scope {
	var values []interface{}
	var orQueryList []string
	for field, value := range opts {
		orQueryList = append(orQueryList, fmt.Sprintf("(%s = ?)", field))
		values = append(values, value)
	}

	s.DB.Where(strings.Join(orQueryList, " OR "), values...)
	return s
}
func (s *Scope) MultiOrLike(opts map[string]string, isPrefix bool) *Scope {
	var values []interface{}
	var orQueryList []string
	for field, value := range opts {
		orQueryList = append(orQueryList, fmt.Sprintf("(%s LIKE ?)", field))
		if isPrefix {
			values = append(values, value+"%")
		} else {
			values = append(values, "%"+value+"%")
		}
	}

	s.DB.Where(strings.Join(orQueryList, " OR "), values...)
	return s
}

// Order db.Order("name DESC")
func (s *Scope) Order(value string) *Scope {
	s.DB = s.DB.Order(value)
	return s
}

func (s *Scope) Orders(values ...OrderColumn) *Scope {
	if len(values) == 0 {
		return s
	}
	var columns []clause.OrderByColumn
	for _, value := range values {
		columns = append(columns, clause.OrderByColumn{
			Column:  clause.Column{Name: value.Field},
			Desc:    value.Desc,
			Reorder: value.Reorder,
		})
	}
	s.DB.Statement.AddClause(clause.OrderBy{
		Columns: columns,
	})
	return s
}

func (s *Scope) Group(name string) *Scope {
	s.DB.Group(name)
	return s
}

func (s *Scope) Groups(names ...string) *Scope {
	var columns []clause.Column
	for _, name := range names {
		fields := strings.FieldsFunc(name, utils.IsValidDBNameChar)
		columns = append(columns, clause.Column{Name: name, Raw: len(fields) != 1})
	}
	s.DB.Statement.AddClause(clause.GroupBy{
		Columns: columns,
	})
	return s
}

func (s *Scope) Having(query interface{}, args ...interface{}) *Scope {
	s.DB.Having(query, args...)
	return s
}
func (s *Scope) Joins(query string, args ...interface{}) *Scope {
	s.DB.Joins(query, args...)
	return s
}

func (s *Scope) Count() (total int64, err error) {
	err = s.DB.Count(&total).Error
	return total, err
}

func (s *Scope) Offset(offset int) *Scope {
	s.DB = s.DB.Offset(offset)
	return s
}

func (s *Scope) Limit(limit int) *Scope {
	s.DB = s.DB.Limit(limit)
	return s
}

func (s *Scope) Omit(value ...string) *Scope {
	s.DB.Omit(value...)
	return s
}

func (s *Scope) Unscoped() *Scope {
	s.DB.Unscoped()
	return s
}

func (s *Scope) Preload(query string, args ...interface{}) *Scope {
	s.DB.Preload(query, args...)
	return s
}

func (s *Scope) Association(value string) *Scope {
	s.DB.Association(value)
	return s
}
