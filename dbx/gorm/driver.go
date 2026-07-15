package gorm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/utils"

	"github.com/ml444/gkit/dbx"
)

// Driver implements dbx.Driver using GORM.
type Driver struct {
	db *gorm.DB
}

func (d *Driver) WithContext(ctx context.Context) dbx.Driver {
	return &Driver{db: d.db.WithContext(ctx)}
}

func applyTxOptions(opts []dbx.TxOption) sql.TxOptions {
	var o sql.TxOptions
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

func (d *Driver) applyBuilder(b *dbx.QueryBuilder) *gorm.DB {
	tx := d.db
	if b.Model != nil {
		tx = tx.Model(b.Model)
	} else if b.Table != "" {
		tx = tx.Table(b.Table)
	}
	if b.Unscoped {
		tx = tx.Unscoped()
	}
	if b.Distinct {
		tx = tx.Distinct()
	}
	if len(b.Selects) > 0 {
		tx = tx.Select(b.Selects)
	}
	if len(b.Omits) > 0 {
		tx = tx.Omit(b.Omits...)
	}
	for _, w := range b.Wheres {
		tx = tx.Where(w.Query, w.Args...)
	}
	for _, w := range b.OrWheres {
		tx = tx.Or(w.Query, w.Args...)
	}
	for _, j := range b.Joins {
		tx = tx.Joins(j.Query, j.Args...)
	}
	if len(b.Groups) > 0 {
		var columns []clause.Column
		for _, name := range b.Groups {
			fields := stringsFieldsFunc(name, utils.IsInvalidDBNameChar)
			columns = append(columns, clause.Column{Name: name, Raw: len(fields) != 1})
		}
		tx.Statement.AddClause(clause.GroupBy{Columns: columns})
	}
	if b.Having != nil {
		tx = tx.Having(b.Having.Query, b.Having.Args...)
	}
	if len(b.Orders) > 0 {
		var columns []clause.OrderByColumn
		for _, value := range b.Orders {
			columns = append(columns, clause.OrderByColumn{
				Column:  clause.Column{Name: value.Field},
				Desc:    value.Desc,
				Reorder: value.Reorder,
			})
		}
		tx.Statement.AddClause(clause.OrderBy{Columns: columns})
	}
	for _, raw := range b.OrderRaw {
		tx = tx.Order(raw)
	}
	if b.Limit > 0 {
		tx = tx.Limit(b.Limit)
	}
	if b.Offset > 0 {
		tx = tx.Offset(b.Offset)
	}
	if b.ForUpdate {
		tx = tx.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if len(b.ReturningColumns) > 0 {
		var cols []clause.Column
		for _, col := range b.ReturningColumns {
			cols = append(cols, clause.Column{Name: col})
		}
		tx = tx.Clauses(clause.Returning{Columns: cols})
	}
	return tx
}

func stringsFieldsFunc(s string, fn func(r rune) bool) []string {
	return stringsSplitFieldsFunc(s, fn)
}

// avoid importing strings in many places - use small helper
func stringsSplitFieldsFunc(s string, fn func(r rune) bool) []string {
	var parts []string
	start := -1
	for i, r := range s {
		if fn(r) {
			if start >= 0 {
				parts = append(parts, s[start:i])
				start = -1
			}
			continue
		}
		if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		parts = append(parts, s[start:])
	}
	return parts
}

func (d *Driver) Find(ctx context.Context, b *dbx.QueryBuilder, dest any) error {
	tx := d.applyBuilder(b)
	return mapErr(tx.Find(dest).Error)
}

func (d *Driver) First(ctx context.Context, b *dbx.QueryBuilder, dest any) error {
	tx := d.applyBuilder(b)
	return mapErr(tx.First(dest).Error)
}

func (d *Driver) Count(ctx context.Context, b *dbx.QueryBuilder) (int64, error) {
	tx := d.applyBuilder(b)
	var total int64
	err := tx.Count(&total).Error
	return total, err
}

func (d *Driver) Create(ctx context.Context, b *dbx.QueryBuilder, v any) (int64, error) {
	tx := d.applyBuilder(b)
	tx = tx.Create(v)
	return tx.RowsAffected, tx.Error
}

func (d *Driver) CreateInBatches(ctx context.Context, b *dbx.QueryBuilder, values any, batchSize int) (int64, error) {
	tx := d.applyBuilder(b).CreateInBatches(values, batchSize)
	return tx.RowsAffected, tx.Error
}

func (d *Driver) Save(ctx context.Context, b *dbx.QueryBuilder, v any) (int64, error) {
	tx := d.applyBuilder(b).Save(v)
	return tx.RowsAffected, tx.Error
}

func (d *Driver) Update(ctx context.Context, b *dbx.QueryBuilder, v any) (int64, error) {
	tx := d.applyBuilder(b).Updates(v)
	return tx.RowsAffected, tx.Error
}

func (d *Driver) UpdateColumn(ctx context.Context, b *dbx.QueryBuilder, field string, value any) (int64, error) {
	tx := d.applyBuilder(b)
	if b.IncrColumn != "" && b.IncrValue != 0 {
		col := b.IncrColumn
		v := b.IncrValue
		if v > 0 {
			tx = tx.UpdateColumn(col, gorm.Expr(fmt.Sprintf("COALESCE(%s, 0) + ?", col), v))
		} else {
			tx = tx.UpdateColumn(col, gorm.Expr(fmt.Sprintf("COALESCE(%s, 0) - ?", col), -v))
		}
	} else {
		tx = tx.UpdateColumn(field, value)
	}
	return tx.RowsAffected, tx.Error
}

func (d *Driver) Delete(ctx context.Context, b *dbx.QueryBuilder) (int64, error) {
	tx := d.applyBuilder(b)
	if b.Model != nil {
		tx = tx.Delete(b.Model)
	} else {
		tx = tx.Delete(nil)
	}
	return tx.RowsAffected, tx.Error
}

func (d *Driver) Scan(ctx context.Context, b *dbx.QueryBuilder, dest any) error {
	tx := d.applyBuilder(b)
	return mapErr(tx.Scan(dest).Error)
}

func (d *Driver) Transaction(ctx context.Context, fn func(dbx.Driver) error, opts ...dbx.TxOption) error {
	var o sql.TxOptions
	for _, opt := range opts {
		opt(&o)
	}
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&Driver{db: tx})
	}, &o)
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dbx.ErrRecordNotFound
	}
	return err
}

// RawDB returns the underlying GORM handle for a driver.
func RawDB(d dbx.Driver) (*gorm.DB, bool) {
	gd, ok := d.(*Driver)
	if !ok {
		if td, ok := d.(*txDriver); ok {
			return td.db, true
		}
		return nil, false
	}
	return gd.db, true
}
