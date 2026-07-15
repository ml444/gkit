package dbx

import (
	"context"
	"fmt"
)

type testRow struct {
	ID        int64  `json:"id" gorm:"primaryKey"`
	Name      string `json:"name"`
	Age       int64  `json:"age"`
	DeletedAt uint32 `json:"deleted_at"`
}

func (testRow) TableName() string { return "test_rows" }

type encryptRow struct {
	ID   uint64 `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"column:name;encrypt:true"`
}

func (encryptRow) TableName() string { return "encrypt_rows" }

type prefixCipher struct{}

func (prefixCipher) Encrypt(v any) (any, error) { return "enc:" + fmt.Sprint(v), nil }
func (prefixCipher) Decrypt(v any) (any, error) {
	s, ok := v.(string)
	if !ok || len(s) < 4 || s[:4] != "enc:" {
		return nil, fmt.Errorf("invalid cipher value %v", v)
	}
	return s[4:], nil
}

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
