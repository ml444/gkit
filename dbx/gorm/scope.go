package gorm

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/ml444/gkit/dbx"
	"github.com/ml444/gkit/dbx/pagination"
)

// GormScope extends dbx.Scope with GORM-specific chain methods.
type GormScope struct {
	*dbx.Scope
	gormDB *gorm.DB
}

// AsGormScope wraps a dbx.Scope when backed by GORM.
func AsGormScope(s *dbx.Scope) (*GormScope, bool) {
	if s == nil || s.Conn() == nil {
		return nil, false
	}
	conn, ok := s.Conn().(*Conn)
	if !ok {
		return nil, false
	}
	d := conn.Driver(s.Context())
	gdb, ok := RawDB(d)
	if !ok {
		return nil, false
	}
	gdb = applyBuilder(gdb, s.Builder())
	return &GormScope{Scope: s, gormDB: gdb}, true
}

// TxConn wraps a transaction *gorm.DB as dbx.Conn for NewScope inside tx callbacks.
func TxConn(tx *gorm.DB) dbx.Conn {
	if tx == nil {
		return nil
	}
	return dbx.StaticConn(&Driver{db: tx})
}

// NewGormScope builds a GormScope from *gorm.DB (including transaction handles).
func NewGormScope(tx *gorm.DB, modelOrTable any) *GormScope {
	s := dbx.NewScope(TxConn(tx), modelOrTable)
	return WrapScope(s, tx)
}

// WrapScope builds GormScope from scope and explicit gorm handle.
func WrapScope(s *dbx.Scope, db *gorm.DB) *GormScope {
	if s == nil {
		return nil
	}
	gdb := applyBuilder(db, s.Builder())
	return &GormScope{Scope: s, gormDB: gdb}
}

func applyBuilder(db *gorm.DB, b *dbx.QueryBuilder) *gorm.DB {
	return (&Driver{db: db}).applyBuilder(b)
}

func (gs *GormScope) Preload(query string, args ...any) *GormScope {
	gs.gormDB = gs.gormDB.Preload(query, args...)
	return gs
}

func (gs *GormScope) Association(name string) *gorm.Association {
	return gs.gormDB.Association(name)
}

func (gs *GormScope) Clauses(exprs ...clause.Expression) *GormScope {
	gs.gormDB = gs.gormDB.Clauses(exprs...)
	return gs
}

func (gs *GormScope) Unscoped() *GormScope {
	ns := gs.Scope.SetIncludeDeleted()
	gdb := gs.gormDB.Unscoped()
	return &GormScope{Scope: ns, gormDB: gdb}
}

// PaginationQueryWithOpt runs paginated find, using deferred join when eligible.
func (gs *GormScope) PaginationQueryWithOpt(list any, opt *pagination.Pagination) (*pagination.Pagination, error) {
	if gs == nil || gs.Scope == nil {
		return nil, errors.New("invalid scope")
	}
	if opt == nil {
		opt = pagination.NewDefaultPagination()
	}
	normalizePagination(opt)

	var total int64
	if !opt.SkipCount {
		var err error
		total, err = gs.Scope.Count()
		if err != nil {
			return nil, err
		}
	}

	offset := opt.Offset()
	limit := int(opt.Size)
	var err error
	if pk, orderExpr, ok := gs.canDeferredJoin(offset); ok {
		err = gs.findWithDeferredJoin(list, limit, offset, pk, orderExpr)
	} else {
		err = gs.gormDB.Limit(limit).Offset(offset).Find(list).Error
	}
	if err != nil {
		return nil, err
	}

	return &pagination.Pagination{
		Page:      opt.Page,
		Size:      opt.Size,
		Total:     total,
		SkipCount: opt.SkipCount,
	}, nil
}

func normalizePagination(opt *pagination.Pagination) {
	if opt.Page == 0 {
		opt.Page = 1
	}
	if opt.Size == 0 {
		opt.Size = uint32(dbx.DefaultLimit)
	} else if opt.Size > uint32(dbx.MaxLimit) {
		opt.Size = uint32(dbx.MaxLimit)
	}
}
