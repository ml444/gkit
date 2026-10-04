package dbx

import (
	"context"
	"database/sql"
	"testing"
)

// stubDriver is a no-op Driver for connection-helper unit tests (avoid importing dbx/mock).
type stubDriver struct{}

func (stubDriver) Find(ctx context.Context, b *QueryBuilder, dest any) error { return nil }
func (stubDriver) First(ctx context.Context, b *QueryBuilder, dest any) error {
	return ErrRecordNotFound
}
func (stubDriver) Count(ctx context.Context, b *QueryBuilder) (int64, error) { return 0, nil }
func (stubDriver) Create(ctx context.Context, b *QueryBuilder, v any) (int64, error) {
	return 1, nil
}
func (stubDriver) CreateInBatches(ctx context.Context, b *QueryBuilder, values any, batchSize int) (int64, error) {
	return 1, nil
}
func (stubDriver) Save(ctx context.Context, b *QueryBuilder, v any) (int64, error) { return 1, nil }
func (stubDriver) Update(ctx context.Context, b *QueryBuilder, v any) (int64, error) {
	return 1, nil
}
func (stubDriver) UpdateColumn(ctx context.Context, b *QueryBuilder, field string, value any) (int64, error) {
	return 1, nil
}
func (stubDriver) Delete(ctx context.Context, b *QueryBuilder) (int64, error) { return 1, nil }
func (stubDriver) Scan(ctx context.Context, b *QueryBuilder, dest any) error {
	return ErrRecordNotFound
}
func (stubDriver) Transaction(ctx context.Context, fn func(Driver) error, opts ...TxOption) error {
	return fn(stubDriver{})
}
func (stubDriver) WithContext(ctx context.Context) Driver { return stubDriver{} }

type stubTxConn struct{ d Driver }

func (c stubTxConn) Driver(ctx context.Context) Driver                           { return c.d }
func (c stubTxConn) Begin(ctx context.Context, opts ...TxOption) (Driver, error) { return c.d, nil }
func (c stubTxConn) Commit(d Driver) error                                       { return nil }
func (c stubTxConn) Rollback(d Driver) error                                     { return nil }


type noContextDriver struct{}

func (noContextDriver) Find(ctx context.Context, b *QueryBuilder, dest any) error {
	return stubDriver{}.Find(ctx, b, dest)
}
func (noContextDriver) First(ctx context.Context, b *QueryBuilder, dest any) error {
	return stubDriver{}.First(ctx, b, dest)
}
func (noContextDriver) Count(ctx context.Context, b *QueryBuilder) (int64, error) {
	return stubDriver{}.Count(ctx, b)
}
func (noContextDriver) Create(ctx context.Context, b *QueryBuilder, v any) (int64, error) {
	return stubDriver{}.Create(ctx, b, v)
}
func (noContextDriver) CreateInBatches(ctx context.Context, b *QueryBuilder, values any, batchSize int) (int64, error) {
	return stubDriver{}.CreateInBatches(ctx, b, values, batchSize)
}
func (noContextDriver) Save(ctx context.Context, b *QueryBuilder, v any) (int64, error) {
	return stubDriver{}.Save(ctx, b, v)
}
func (noContextDriver) Update(ctx context.Context, b *QueryBuilder, v any) (int64, error) {
	return stubDriver{}.Update(ctx, b, v)
}
func (noContextDriver) UpdateColumn(ctx context.Context, b *QueryBuilder, field string, value any) (int64, error) {
	return stubDriver{}.UpdateColumn(ctx, b, field, value)
}
func (noContextDriver) Delete(ctx context.Context, b *QueryBuilder) (int64, error) {
	return stubDriver{}.Delete(ctx, b)
}
func (noContextDriver) Scan(ctx context.Context, b *QueryBuilder, dest any) error {
	return stubDriver{}.Scan(ctx, b, dest)
}
func (noContextDriver) Transaction(ctx context.Context, fn func(Driver) error, opts ...TxOption) error {
	return fn(noContextDriver{})
}

func TestDriverConnectionHelpers(t *testing.T) {
	d := stubDriver{}
	conn := stubTxConn{d: d}
	if StaticConn(d).Driver(context.Background()) != d {
		t.Fatal("static connection returned wrong driver")
	}
	if _, ok := AsTxManager(conn); !ok {
		t.Fatal("tx connection should be a TxManager")
	}
	if _, ok := AsTxManager(StaticConn(d)); ok {
		t.Fatal("static connection should not be a TxManager")
	}
	want := sql.TxOptions{Isolation: sql.LevelSerializable, ReadOnly: true}
	var got sql.TxOptions
	WithTxOptions(want)(&got)
	if got != want {
		t.Fatalf("options = %#v, want %#v", got, want)
	}
	if StaticConn(d).Driver(context.WithValue(context.Background(), "key", "value")) != d {
		t.Fatal("WithContext-capable driver should be returned")
	}
	plain := noContextDriver{}
	if StaticConn(plain).Driver(context.Background()) != plain {
		t.Fatal("plain static driver should be returned unchanged")
	}
	if got := wrapTxConn(StaticConn(plain), plain).Driver(context.Background()); got != plain {
		t.Fatal("transaction connection should return plain tx driver")
	}
	if got := wrapTxConn(StaticConn(d), d).Driver(context.Background()); got != d {
		t.Fatal("transaction connection should bind context-capable tx driver")
	}
	if got := wrapTxConn(StaticConn(plain), nil).Driver(context.Background()); got != plain {
		t.Fatal("transaction connection should fall back to base connection")
	}
}
