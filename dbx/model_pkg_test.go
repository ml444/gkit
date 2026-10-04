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
	batch := []*testRow{{ID: 2, Name: "bob"}, {ID: 3, Name: "cat"}}
	if err := repo.BatchCreate(ctx, &batch, "age"); err != nil {
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

type srcModel struct {
	ID     uint64 `json:"id"`
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

func (m *srcModel) ToORM() dbx.ITModel {
	return &ormModel{ID: m.ID, Name: m.Name, Secret: m.Secret, force: true}
}

type ormModel struct {
	ID   uint64 `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
	Secret string `json:"secret"`

	force bool
}

func (ormModel) TableName() string { return "orm_models" }

func (o *ormModel) ToSource() dbx.IModel {
	return &srcModel{ID: o.ID, Name: o.Name}
}

func (o *ormModel) ForceTModel() bool { return o.force }

func (o *ormModel) CopyToSource(dst dbx.IModel) error {
	m := dst.(*srcModel)
	m.ID, m.Name = o.ID, o.Name
	return nil
}

func (o *ormModel) CopyToSourceIgnoreEmpty(dst dbx.IModel) error {
	m := dst.(*srcModel)
	if o.ID != 0 {
		m.ID = o.ID
	}
	if o.Name != "" {
		m.Name = o.Name
	}
	return nil
}

func TestRepositoryCopiesForcedORMModels(t *testing.T) {
	d, conn := testConn()
	repo := dbx.NewT[srcModel](func() dbx.Conn { return conn })
	ctx := context.Background()

	created := &srcModel{ID: 1, Name: "alice"}
	if err := repo.Create(ctx, created); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got := d.Tables["orm_models"][0]["name"]; got != "alice" {
		t.Fatalf("stored name = %v", got)
	}

	var one srcModel
	if err := repo.GetOne(ctx, &one, uint64(1)); err != nil {
		t.Fatalf("GetOne: %v", err)
	}
	if one != (srcModel{ID: 1, Name: "alice"}) {
		t.Fatalf("GetOne result = %#v", one)
	}

	list := []srcModel{{Name: "keep"}}
	if err := repo.ListAll(ctx, &list, nil); err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(list) != 1 || list[0] != (srcModel{ID: 1, Name: "alice"}) {
		t.Fatalf("ListAll result = %#v", list)
	}

	batch := []*srcModel{{ID: 2, Name: "bob"}, {ID: 3, Name: "cara"}}
	if err := repo.BatchCreate(ctx, &batch); err != nil {
		t.Fatalf("BatchCreate: %v", err)
	}
	if len(d.Tables["orm_models"]) != 3 || batch[1].Name != "cara" {
		t.Fatalf("BatchCreate result = %#v", batch)
	}

	if rows, err := repo.Update(ctx, &srcModel{Name: "updated"}, "id = ?", uint64(1)); err != nil || rows != 1 {
		t.Fatalf("Update = %d, %v", rows, err)
	}
	if rows, err := repo.UpdateByPk(ctx, &srcModel{Name: "again"}, uint64(1)); err != nil || rows != 1 {
		t.Fatalf("UpdateByPk = %d, %v", rows, err)
	}
	if err := repo.Save(ctx, &srcModel{ID: 4, Name: "saved"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	var paged []srcModel
	if _, err := repo.ListWithPagination(ctx, &paged, nil, 1, 2); err != nil || len(paged) != 2 {
		t.Fatalf("ListWithPagination = %#v, %v", paged, err)
	}
	if count, err := repo.Count(ctx, "id = ?", uint64(1)); err != nil || count != 1 {
		t.Fatalf("Count query = %d, %v", count, err)
	}
	if err := repo.DeleteByPk(ctx, uint64(1)); err != nil {
		t.Fatalf("DeleteByPk: %v", err)
	}
	if err := repo.DeleteByWhere(ctx, "id = ?", uint64(2)); err != nil {
		t.Fatalf("DeleteByWhere: %v", err)
	}
}
