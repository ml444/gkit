package dbx

import (
	"errors"

	"github.com/ml444/gkit/dbx/pagination"
	"github.com/ml444/gkit/log"
)

const (
	DefaultLimit uint32 = 2000
	MaxLimit     uint32 = 100000
)

func (s *Scope) PaginationQuery(opt *pagination.Pagination, list interface{}) (*pagination.Pagination, error) {
	var err error
	opt = s.HandlePagination(opt)
	offset := opt.Offset()

	var total int64
	if offset == 0 || !opt.SkipCount {
		err = s.DB.Count(&total).Error
		if err != nil {
			return nil, err
		}
	}
	if s == nil || s.DB == nil {
		return nil, errors.New("invalid scope or transaction")
	}
	s.DB = s.DB.Limit(int(opt.Size)).Offset(offset)
	err = s.Find(list)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	p := pagination.Pagination{
		Page:  opt.Page,
		Size:  opt.Size,
		Total: total,
	}
	return &p, nil
}

func (s *Scope) HandlePagination(opt *pagination.Pagination) *pagination.Pagination {
	if opt == nil {
		opt = &pagination.Pagination{}
	}

	if opt.Size == 0 {
		opt.Size = DefaultLimit
	} else if opt.Size > MaxLimit {
		opt.Size = MaxLimit
	}

	return opt
}
