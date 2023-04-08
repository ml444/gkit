package dbx

import (
	"reflect"
	"time"

	"github.com/ml444/gkit/listoption"
	log "github.com/ml444/glog"
	"gorm.io/gorm"
)

const SoftDeleteObjField = "DeletedAt"
const SoftDeleteDbField = "deleted_at"
const DefaultLimit int = 2000
const DefaultOffset int = 0
const MaxLimit int = 100000

type Scope struct {
	tx           *gorm.DB
	model        interface{}
	listOption   *listoption.ListOption
	SQL          string
	RowsAffected int64
}

func NewScope(db *gorm.DB, model interface{}) *Scope {
	return &Scope{
		tx:    db.Model(model),
		model: model,
	}
}

func (s *Scope) fillResult() {
	s.SQL = s.tx.Statement.SQL.String()
	s.RowsAffected = s.tx.RowsAffected
}

func (s *Scope) Create(v interface{}) error {
	return s.tx.Create(v).Error
}

// CreateInBatches Insert data in batches after splitting data according to batchSize
func (s *Scope) CreateInBatches(values interface{}, batchSize int) error {
	defer s.fillResult()
	return s.tx.CreateInBatches(values, batchSize).Error
}

// Update updates attributes using callbacks. values must be a struct or map.
func (s *Scope) Update(v interface{}, conds ...interface{}) error {
	if condsLen := len(conds); condsLen == 1 {
		s.tx.Where(conds[0])
	} else if condsLen > 1 {
		s.tx.Where(conds[0], conds[1:])
	}
	s.tx.Updates(v)
	if s.tx.Error != nil {
		return s.tx.Error
	}
	if s.tx.RowsAffected == 0 {
		log.Warnf("SQL: %s; RowsAffected: 0", s.tx.Statement.SQL.String())
	}
	s.fillResult()
	return nil
}

func (s *Scope) Delete(v interface{}, conds ...interface{}) error {
	defer s.fillResult()
	if _, ok := reflect.TypeOf(v).Elem().FieldByName(SoftDeleteObjField); ok {
		if condsLen := len(conds); condsLen == 1 {
			s.tx.Where(conds[0])
		} else if condsLen > 1 {
			s.tx.Where(conds[0], conds[1:])
		}
		return s.tx.UpdateColumn(SoftDeleteDbField, time.Now().Unix()).Error
	}
	return s.tx.Delete(v, conds).Error
}

func (s *Scope) Where(query interface{}, args ...interface{}) *Scope {
	s.tx.Where(query, args)
	return s
}

func (s *Scope) First(dest interface{}, conds ...interface{}) error {
	return s.tx.First(dest, conds).Error
}

func (s *Scope) Find(dest interface{}, conds ...interface{}) error {
	return s.tx.Find(dest, conds).Error
}

// Order specify order when retrieving records from database
//
//	db.Order("name DESC")
//	db.Order(clause.OrderByColumn{Column: clause.Column{Name: "name"}, Desc: true})
func (s *Scope) Order(value interface{}) {
	s.tx = s.tx.Order(value)
}

func (s *Scope) Offset(offset int) {
	s.tx = s.tx.Offset(offset)
}

func (s *Scope) Limit(limit int) {
	s.tx = s.tx.Limit(limit)
}
func (s *Scope) PaginateQuery(opt *listoption.ListOption, list interface{}) (*listoption.Paginate, error) {
	var total int64
	if opt != nil && opt.Offset == 0 && !opt.SkipCount {
		err := s.tx.Count(&total).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}
	s.SetOffsetAndLimitByListOption(opt)
	err := s.tx.Find(list).Error
	if err != nil {
		return nil, err
	}
	p := listoption.Paginate{
		Offset: opt.Offset,
		Limit:  opt.Limit,
		Total:  total,
	}
	return &p, nil
}

func (s *Scope) SetOffsetAndLimitByListOption(opt *listoption.ListOption) {
	var limit int
	var offset int
	if opt != nil {
		if opt.Limit == 0 {
			opt.Limit = uint32(DefaultLimit)
		}
		offset = int(opt.Offset)
		limit = int(opt.Limit)
	} else {
		offset = DefaultOffset
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
		opt.Limit = uint32(MaxLimit)
	}
	s.tx = s.tx.Limit(limit).Offset(offset)
	return
}
