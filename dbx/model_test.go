package dbx_test

import (
	"context"
	"testing"

	"github.com/ml444/gkit/dbx"
)

func TestRepositoryOperationsAndOptions(t *testing.T) {
	d, conn := testConn()
	repo := dbx.NewT[testRow](func() dbx.Conn { return conn }, dbx.SetCreateBatchSize(2), dbx.SetPrimaryKey("id"), dbx.SetNotFoundErrCode(91))
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
	clone := repo.Clone(dbx.SetIgnoreNotFoundErr(), dbx.SetDisableDecrypt())
	if !clone.IgnoreNotFoundErr || !clone.DisableDecrypt || clone.BatchCreateSize != 2 || clone.PrimaryKey != "id" {
		t.Fatal("clone/options incorrect")
	}
	if d == nil {
		t.Fatal("sanity")
	}
}

func TestRepositoryCrypto(t *testing.T) {
	d, conn := testConn()
	repo := dbx.NewT[encryptRow](func() dbx.Conn { return conn }, dbx.SetTableCipher(prefixCipher{}))
	ctx := context.Background()
	row := &encryptRow{ID: 1, Name: "alice"}
	if err := repo.Create(ctx, row); err != nil || row.Name != "enc:alice" {
		t.Fatalf("encrypted create: %#v %v", row, err)
	}
	var got encryptRow
	if err := repo.GetOne(ctx, &got, uint64(1)); err != nil || got.Name != "alice" {
		t.Fatalf("decrypted get: %#v %v", got, err)
	}
	if d.Tables["encrypt_rows"][0]["name"] != "enc:alice" {
		t.Fatal("stored value was not encrypted")
	}
	value := map[string]any{"name": "bob"}
	if err := repo.CheckAndCrypto(value, dbx.CipherKindEncrypt, false); err != nil || value["name"] != "enc:bob" {
		t.Fatalf("map crypto: %#v %v", value, err)
	}
	list := []*encryptRow{{Name: "cat"}}
	if err := repo.CheckAndCrypto(list, dbx.CipherKindEncrypt, false); err != nil || list[0].Name != "enc:cat" {
		t.Fatalf("slice crypto: %#v %v", list, err)
	}
	specified := dbx.NewT[encryptRow](func() dbx.Conn { return conn }, dbx.SetSpecifyFieldCipherMap(map[string]dbx.FieldCipher{"name": {StructField: "Name", Cipher: prefixCipher{}}}), dbx.SetDisableDecrypt())
	if !specified.NeedEncrypt || !specified.DisableDecrypt {
		t.Fatal("specified cipher options missing")
	}
}
