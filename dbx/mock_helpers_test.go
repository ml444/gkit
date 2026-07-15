package dbx_test

import (
	"fmt"

	"github.com/ml444/gkit/dbx"
	"github.com/ml444/gkit/dbx/mock"
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

func testConn() (*mock.Driver, dbx.Conn) {
	d := mock.New()
	return d, d.Conn()
}
