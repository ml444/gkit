package dbx

import (
	"errors"

	"github.com/ml444/gkit/dbx/pagination"
	"github.com/ml444/gkit/log"
)

const (
	DefaultLimit int = 2000
	MaxLimit     int = 100000
)

func (s *Scope) PaginationQuery(list interface{}, page, size uint32) (*pagination.Pagination, error) {
	var err error
	opt := s.HandlePagination(page, size)
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

func (s *Scope) QueryPagination(list interface{}, page, size uint32, skipCount bool) (total int64, err error) {
	offset := getOffset(page, size)

	if offset == 0 || !skipCount {
		err = s.DB.Count(&total).Error
		if err != nil {
			return
		}
	}
	if s == nil || s.DB == nil {
		return 0, errors.New("invalid scope or transaction")
	}
	s.DB = s.DB.Limit(getLimit(size)).Offset(offset)
	err = s.Find(list)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (s *Scope) HandlePagination(page, size uint32) *pagination.Pagination {
	opt := &pagination.Pagination{
		Page: page,
		Size: size,
	}

	if opt.Size == 0 {
		opt.Size = uint32(DefaultLimit)
	} else if opt.Size > uint32(MaxLimit) {
		opt.Size = uint32(MaxLimit)
	}

	return opt
}

func getLimit(size uint32) int {
	limit := int(size)
	if limit == 0 {
		return DefaultLimit
	} else if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}


func getOffset(page, size uint32) int {
	if page <= 1 {
		return 0
	}
	
return int(size * (page - 1))
}

