package dbx_test

import (
	"context"
	"strings"
	"testing"

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

func testConn() (*mock.Driver, dbx.Conn) {
	d := mock.New()
	return d, d.Conn()
}

func TestRepositoryOperationsAndOptions(t *testing.T) {
	d, conn := testConn()
	repo := dbx.NewT[testRow](func() dbx.Conn { return conn },
		dbx.WithTCreateBatchSize(2),
		dbx.WithTPrimaryKey("id"),
		dbx.WithTNotFoundErrCode(91),
	)
	ctx := context.Background()
	a := &testRow{ID: 1, Name: "alice", Age: 10}
	if err := repo.Create(ctx, a, "age"); err != nil || a.ID != 1 {
		t.Fatalf("create: %#v, %v", a, err)
	}
	if err := repo.BatchCreate(ctx, []*testRow{{ID: 2, Name: "bob"}, {ID: 3, Name: "cat"}}, "age"); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(ctx, &testRow{ID: 4, Name: "dave"}); err != nil {
		t.Fatal(err)
	}
	if rows, err := repo.Update(ctx, map[string]any{"name": "x"}, "id = ?", int64(2)); err != nil || rows != 1 {
		t.Fatalf("update: %d %v", rows, err)
	}
	if rows, err := repo.UpdateByPk(ctx, map[string]any{"name": "y"}, int64(2), "name"); err != nil || rows != 1 {
		t.Fatalf("update pk: %d %v", rows, err)
	}
	if err := repo.DeleteByPk(ctx, int64(2)); err != nil {
		t.Fatal(err)
	}
	if err := repo.DeleteByWhere(ctx, "id = ?", int64(3)); err != nil {
		t.Fatal(err)
	}
	if exists, err := repo.ExistByWhere(ctx, "id = ?", int64(1)); err != nil || !exists {
		t.Fatalf("exist: %v %v", exists, err)
	}
	if count, err := repo.Count(ctx); err != nil || count != 4 {
		t.Fatalf("count: %d %v", count, err)
	}
	var one testRow
	if err := repo.GetOne(ctx, &one, int64(1)); err != nil || one.Name != "alice" {
		t.Fatalf("get: %#v %v", one, err)
	}
	if err := repo.GetOneByWhere(ctx, &one, "id = ?", int64(99)); !dbx.IsNotFoundErr(err, 91) {
		t.Fatalf("missing: %v", err)
	}
	var all []testRow
	if err := repo.ListAll(ctx, &all, map[string]any{"id": int64(1)}); err != nil || len(all) != 1 {
		t.Fatalf("list: %#v %v", all, err)
	}
	var page []testRow
	if p, err := repo.ListWithPagination(ctx, &page, nil, 1, 2); err != nil || p.Total != 4 {
		t.Fatalf("page: %#v %v", p, err)
	}
	clone := repo.Clone(dbx.WithTIgnoreNotFoundErr(), dbx.WithTDecrypt(false))
	if !clone.IgnoreNotFoundErr || clone.BatchCreateSize != 2 || clone.PrimaryKey != "id" {
		t.Fatal("clone/options incorrect")
	}
	if d == nil {
		t.Fatal("sanity")
	}
}

func TestRepositoryCrypto(t *testing.T) {
	d, conn := testConn()
	repo := dbx.NewT[encryptRow](func() dbx.Conn { return conn }, encryptRowCrypto(crypto.MutateInPlace))
	ctx := context.Background()
	row := &encryptRow{ID: 1, Name: "alice"}
	if err := repo.Create(ctx, row); err != nil {
		t.Fatalf("encrypted create: %#v %v", row, err)
	}
	if !strings.HasPrefix(row.Name, "v1:k") {
		t.Fatalf("create should encrypt in place: %#v", row)
	}
	var got encryptRow
	if err := repo.GetOne(ctx, &got, uint64(1)); err != nil || got.Name != "alice" {
		t.Fatalf("decrypted get: %#v %v", got, err)
	}
	if stored, ok := d.Tables["encrypt_rows"][0]["name"].(string); !ok || !strings.HasPrefix(stored, "v1:k") {
		t.Fatalf("stored value was not encrypted: %#v", d.Tables["encrypt_rows"][0]["name"])
	}

	noDecrypt := dbx.NewT[encryptRow](func() dbx.Conn { return conn },
		encryptRowCrypto(),
		dbx.WithTDecrypt(false),
	)
	if err := noDecrypt.GetOne(ctx, &encryptRow{}, uint64(1)); err != nil {
		t.Fatal(err)
	}
}
