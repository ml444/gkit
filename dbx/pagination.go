package dbx

import (
	"errors"

	"github.com/ml444/gkit/dbx/paging"
	"github.com/ml444/gkit/log"
)

const (
	DefaultLimit uint32 = 2000
	MaxLimit     uint32 = 100000
)

func (s *Scope) PaginateQuery(opt *paging.Paginate, list interface{}) (*paging.Paginate, error) {
	var err error
	opt = s.HandlePaginate(opt)
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
	p := paging.Paginate{
		Page:  opt.Page,
		Size:  opt.Size,
		Total: total,
	}
	return &p, nil
}

func (s *Scope) HandlePaginate(opt *paging.Paginate) *paging.Paginate {
	if opt == nil {
		opt = &paging.Paginate{}
	}

	if opt.Size == 0 {
		opt.Size = DefaultLimit
	} else if opt.Size > MaxLimit {
		opt.Size = MaxLimit
	}

	return opt
}
