package dbx

import (
	"context"
	"database/sql"
)

// Conn provides context-scoped Driver instances.
type Conn interface {
	Driver(ctx context.Context) Driver
}

// Driver executes queries compiled from QueryBuilder.
type Driver interface {
	Find(ctx context.Context, b *QueryBuilder, dest any) error
	First(ctx context.Context, b *QueryBuilder, dest any) error
	Count(ctx context.Context, b *QueryBuilder) (int64, error)
	Create(ctx context.Context, b *QueryBuilder, v any) (int64, error)
	CreateInBatches(ctx context.Context, b *QueryBuilder, values any, batchSize int) (int64, error)
	Save(ctx context.Context, b *QueryBuilder, v any) (int64, error)
	Update(ctx context.Context, b *QueryBuilder, v any) (int64, error)
	UpdateColumn(ctx context.Context, b *QueryBuilder, field string, value any) (int64, error)
	Delete(ctx context.Context, b *QueryBuilder) (int64, error)
	Scan(ctx context.Context, b *QueryBuilder, dest any) error

	Transaction(ctx context.Context, fn func(Driver) error, opts ...TxOption) error
}

// TxManager is an optional interface for explicit transaction control.
type TxManager interface {
	Begin(ctx context.Context, opts ...TxOption) (Driver, error)
	Commit(d Driver) error
	Rollback(d Driver) error
}

// AsTxManager returns TxManager when conn supports it.
func AsTxManager(conn Conn) (TxManager, bool) {
	m, ok := conn.(TxManager)
	return m, ok
}

// TxOption configures transactions.
type TxOption func(*sql.TxOptions)

// WithTxOptions sets sql.TxOptions for Transaction/Begin.
func WithTxOptions(opts sql.TxOptions) TxOption {
	return func(o *sql.TxOptions) {
		*o = opts
	}
}

type staticConn struct {
	driver Driver
}

func (c *staticConn) Driver(ctx context.Context) Driver {
	if wd, ok := c.driver.(contextDriver); ok {
		return wd.WithContext(ctx)
	}
	return c.driver
}

// StaticConn wraps a Driver as Conn (for transaction-scoped drivers).
func StaticConn(d Driver) Conn {
	return &staticConn{driver: d}
}

type contextDriver interface {
	WithContext(ctx context.Context) Driver
}

type txConn struct {
	base Conn
	tx   Driver
}

func (c *txConn) Driver(ctx context.Context) Driver {
	if c.tx != nil {
		if wd, ok := c.tx.(contextDriver); ok {
			return wd.WithContext(ctx)
		}
		return c.tx
	}
	return c.base.Driver(ctx)
}

func wrapTxConn(base Conn, tx Driver) Conn {
	return &txConn{base: base, tx: tx}
}
