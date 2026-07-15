package gorm

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	"github.com/ml444/gkit/dbx"
)

// Conn wraps *gorm.DB as dbx.Conn.
type Conn struct {
	db *gorm.DB
}

// NewConn returns a dbx.Conn backed by GORM.
func NewConn(db *gorm.DB) dbx.Conn {
	return &Conn{db: db}
}

func (c *Conn) Driver(ctx context.Context) dbx.Driver {
	return &Driver{db: c.db.WithContext(ctx)}
}

// DB returns the underlying *gorm.DB.
func (c *Conn) DB() *gorm.DB {
	return c.db
}

// Begin starts an explicit transaction (dbx.TxManager).
func (c *Conn) Begin(ctx context.Context, opts ...dbx.TxOption) (dbx.Driver, error) {
	var o sql.TxOptions
	for _, opt := range opts {
		opt(&o)
	}
	tx := c.db.WithContext(ctx).Begin(&o)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &txDriver{Driver: &Driver{db: tx}, tx: tx}, nil
}

func (c *Conn) Commit(d dbx.Driver) error {
	td, ok := d.(*txDriver)
	if !ok {
		return gorm.ErrInvalidTransaction
	}
	return td.tx.Commit().Error
}

func (c *Conn) Rollback(d dbx.Driver) error {
	td, ok := d.(*txDriver)
	if !ok {
		return gorm.ErrInvalidTransaction
	}
	return td.tx.Rollback().Error
}

type txDriver struct {
	*Driver
	tx *gorm.DB
}
