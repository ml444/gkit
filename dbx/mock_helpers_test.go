package dbx_test

import (
	"github.com/ml444/gkit/dbx"
	"github.com/ml444/gkit/dbx/crypto"
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
	Name string `json:"name" gorm:"column:name;encrypt:searchable"`
}

func (encryptRow) TableName() string { return "encrypt_rows" }

type storageEncryptRow struct {
	ID     uint64 `json:"id" gorm:"primaryKey"`
	Name   string `json:"name" gorm:"column:name;encrypt:searchable"`
	Secret string `json:"secret" gorm:"column:secret;encrypt:storage"`
}

func (storageEncryptRow) TableName() string { return "storage_encrypt_rows" }

func testKeyRing() crypto.KeyRing {
	return crypto.KeyRing{
		ActiveID: "1",
		Keys: map[string][]byte{
			"1": []byte("0123456789abcdef0123456789abcdef"),
		},
	}
}

func encryptRowCrypto(mutate ...crypto.MutatePolicy) dbx.T2Option {
	policy := crypto.MutateInPlace
	if len(mutate) > 0 {
		policy = mutate[0]
	}
	return dbx.WithTCrypto(crypto.Config{
		KeyRing: testKeyRing(),
		Fields:  []crypto.FieldSpec{{Column: "name", Struct: "Name", Mode: crypto.ModeSearchable}},
		Mutate:  policy,
	})
}

func dualFieldCrypto() dbx.T2Option {
	return dbx.WithTCrypto(crypto.Config{
		KeyRing: testKeyRing(),
		Fields: []crypto.FieldSpec{
			{Column: "name", Struct: "Name", Mode: crypto.ModeSearchable},
			{Column: "secret", Struct: "Secret", Mode: crypto.ModeStorage},
		},
		Mutate: crypto.MutateInPlace,
	})
}

func testConn() (*mock.Driver, dbx.Conn) {
	d := mock.New()
	return d, d.Conn()
}
