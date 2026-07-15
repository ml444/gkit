package dbx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ml444/gkit/dbx"
	"github.com/ml444/gkit/dbx/mock"
)

type failCipher struct{ err error }

func (c failCipher) Encrypt(v any) (any, error) {
	if c.err != nil {
		return nil, c.err
	}
	return "x:" + v.(string), nil
}
func (c failCipher) Decrypt(v any) (any, error) {
	if c.err != nil {
		return nil, c.err
	}
	s, _ := v.(string)
	if len(s) > 2 && s[:2] == "x:" {
		return s[2:], nil
	}
	return v, nil
}

func TestModelErrorAndOptionBranches(t *testing.T) {
	_, conn := testConn()
	ctx := context.Background()

	noPK := dbx.NewT[testRow](func() dbx.Conn { return conn }, dbx.SetPrimaryKey(""))
	if _, err := noPK.UpdateByPk(ctx, map[string]any{"name": "x"}, int64(1)); err == nil {
		t.Fatal("UpdateByPk empty pk")
	}
	if err := noPK.DeleteByPk(ctx, int64(1)); err == nil {
		t.Fatal("DeleteByPk empty pk")
	}
	if err := noPK.GetOne(ctx, &testRow{}, int64(1)); err == nil {
		t.Fatal("GetOne empty pk")
	}

	repo := dbx.NewT[encryptRow](func() dbx.Conn { return conn },
		dbx.SetSpecifyFieldCipherMap(map[string]dbx.FieldCipher{"name": {StructField: "Name", Cipher: failCipher{err: errors.New("boom")}}}),
		dbx.SetGenerateIDFunc(func() uint64 { return 42 }),
	)
	if err := repo.Create(ctx, &encryptRow{Name: "a"}); err == nil {
		t.Fatal("Create encrypt error")
	}
	if err := repo.DeleteByWhere(ctx, map[string]any{"name": "a"}); err == nil {
		t.Fatal("DeleteByWhere encrypt error")
	}
	if _, err := repo.ExistByWhere(ctx, map[string]any{"name": "a"}); err == nil {
		t.Fatal("ExistByWhere encrypt error")
	}
	if _, err := repo.Count(ctx, map[string]any{"name": "a"}); err == nil {
		t.Fatal("Count encrypt error")
	}

	okRepo := dbx.NewT[encryptRow](func() dbx.Conn { return conn },
		dbx.SetSpecifyFieldCipherMap(map[string]dbx.FieldCipher{"name": {StructField: "Name", Cipher: prefixCipher{}}}),
		dbx.SetGenerateIDFunc(func() uint64 { return 7 }),
	)
	row := &encryptRow{Name: "bob"}
	if err := okRepo.Create(ctx, row); err != nil || row.ID != 7 || row.Name != "enc:bob" {
		t.Fatalf("id generator create: %#v %v", row, err)
	}
	m := map[string]any{"name": "carl"}
	if err := okRepo.CheckAndCrypto(m, dbx.CipherKindEncrypt, true); err != nil || m["id"] == nil {
		t.Fatalf("map id generator: %#v %v", m, err)
	}

	plain := dbx.NewT[testRow](func() dbx.Conn { return conn })
	if _, err := plain.ExistByWhere(ctx); err != nil {
		t.Fatal(err)
	}
	var bad any = 1
	if err := plain.ListAll(ctx, bad, nil); err == nil {
		t.Fatal("ListAll non-pointer")
	}
	var notSlice *int
	if err := plain.ListAll(ctx, notSlice, nil); err == nil {
		t.Fatal("ListAll non-slice")
	}
	var list []testRow
	if err := plain.ListAll(ctx, &list, dbx.QueryOpts{Where: map[string]any{"id": int64(7)}}); err != nil {
		t.Fatal(err)
	}
	if err := plain.ListAll(ctx, &list, &dbx.QueryOpts{Where: map[string]any{"id": int64(7)}}); err != nil {
		t.Fatal(err)
	}
	scope := dbx.NewScope(conn, &testRow{})
	if err := plain.ListAll(ctx, &list, scope); err != nil {
		t.Fatal(err)
	}
	if err := plain.ListAll(ctx, &list, *scope); err != nil {
		t.Fatal(err)
	}
	if err := plain.ListAll(ctx, &list, 123); err == nil {
		t.Fatal("unknown opts")
	}
	if err := plain.ListAll(ctx, &list, ""); err != nil {
		t.Fatal(err)
	}
	exists, err := plain.ExistByWhere(ctx, "id = ?", int64(7))
	if err != nil {
		t.Fatal(err)
	}
	_ = exists

	ignored := plain.Clone(dbx.SetIgnoreNotFoundErr())
	var missing testRow
	if err := ignored.GetOneByWhere(ctx, &missing, "id = ?", int64(99999)); err != nil {
		t.Fatalf("ignore not found: %v", err)
	}

	emptyMap := dbx.NewT[testRow](func() dbx.Conn { return conn }, dbx.SetSpecifyFieldCipherMap(nil))
	_ = emptyMap
}

func TestTransactionItemBranches(t *testing.T) {
	d, conn := testConn()
	d.Seed("test_rows", map[string]any{"id": int64(1), "name": "a", "age": int64(1), "deleted_at": uint32(0)})
	ctx := context.Background()
	repo := dbx.NewT[testRow](func() dbx.Conn { return conn })

	if err := dbx.RunTxItems(ctx, conn); err != nil {
		t.Fatal(err)
	}
	if err := dbx.RunTxItemsWithT(ctx, repo); err != nil {
		t.Fatal(err)
	}

	ins := dbx.NewInsertItem(nil)
	if err := dbx.RunTxItems(ctx, conn, ins); err != nil {
		t.Fatal(err)
	}
	if err := dbx.RunTxItems(ctx, conn, dbx.NewInsertItem(&testRow{ID: 9, Name: "n"})); err != nil {
		t.Fatal(err)
	}

	upd := &dbx.UpdateItem{Model: &testRow{}, Where: map[string]any{"id": int64(1)}, Updates: map[string]any{"name": "u"}}
	if err := dbx.RunTxItems(ctx, conn, upd); err != nil {
		t.Fatal(err)
	}
	updNil := &dbx.UpdateItem{Model: nil, Where: nil}
	_ = updNil.Execute(d)
	updEmpty := &dbx.UpdateItem{Model: &testRow{ID: 1}, Where: map[string]any{"id": int64(1)}}
	if err := updEmpty.Execute(d); err != nil {
		t.Fatal(err)
	}

	save := &dbx.SaveItem{Model: &testRow{ID: 10, Name: "s"}, Where: map[string]any{"id": int64(999)}}
	if err := dbx.RunTxItems(ctx, conn, save); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.SaveItem{}).Preload(d); err == nil {
		t.Fatal("SaveItem nil model")
	}
	_ = (&dbx.SaveItem{Model: &testRow{ID: 11}}).Execute(d)

	si := &dbx.ScopeInsertItem{Models: nil}
	_ = si.Preload(repo, d)
	_ = si.Execute(repo, d)
	if err := dbx.RunTxItemsWithT(ctx, repo, &dbx.ScopeInsertItem{Models: &testRow{ID: 12, Name: "si"}}); err != nil {
		t.Fatal(err)
	}

	su := &dbx.ScopeUpdateItem{Model: &testRow{}, Where: map[string]any{"id": int64(1)}, Updates: map[string]any{"name": "su"}}
	if err := dbx.RunTxItemsWithT(ctx, repo, su); err != nil {
		t.Fatal(err)
	}
	_ = (&dbx.ScopeUpdateItem{Model: nil, Where: nil}).Execute(repo, d)
	su2 := &dbx.ScopeUpdateItem{Model: &testRow{ID: 1}, Where: map[string]any{"id": int64(1)}}
	if err := su2.Execute(repo, d); err != nil {
		t.Fatal(err)
	}

	ss := &dbx.ScopeSaveItem{Model: &testRow{ID: 13, Name: "ss"}, Where: map[string]any{"id": int64(888)}}
	if err := dbx.RunTxItemsWithT(ctx, repo, ss); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.ScopeSaveItem{}).Preload(repo, d); err == nil {
		t.Fatal("ScopeSaveItem nil")
	}
	_ = (&dbx.ScopeSaveItem{Model: &testRow{ID: 14}}).Execute(repo, d)

	if err := dbx.ScopeTxGoWithT(ctx, repo, func() (any, func(*dbx.Scope) error) {
		return nil, func(s *dbx.Scope) error { return s.Create(&testRow{ID: 15, Name: "tx"}) }
	}); err != nil {
		t.Fatal(err)
	}
}

func TestScopeEdgeBranches(t *testing.T) {
	d, conn := testConn()
	d.Seed("test_rows", map[string]any{"id": int64(1), "name": "a", "age": int64(1), "deleted_at": uint32(0)})
	s := dbx.NewScope(conn, &testRow{})
	if err := s.UpdateColumnWithIncr("age", 0); err != nil {
		t.Fatal(err)
	}
	if err := s.Eq("id", int64(99)).UpdateColumn("name", "x"); err != nil {
		t.Fatal(err)
	}
	if err := s.Eq("id", int64(99)).UpdateColumnWithIncr("age", 1); !errors.Is(err, dbx.ErrUpdateRowAffectedZero) {
		t.Fatalf("incr zero rows: %v", err)
	}
	_ = s.Or(map[string]any{"id": int64(1)})
	_ = s.Or("id = ?", int64(1))
	_ = s.Having(map[string]any{"x": 1})
	var rows []testRow
	if err := s.Find(&rows, "id = ?", int64(1)); err != nil {
		t.Fatal(err)
	}
	if b := (*dbx.Scope)(nil).Builder(); b == nil {
		t.Fatal("nil builder")
	}
}

func TestPaginationErrorPaths(t *testing.T) {
	_, conn := testConn()
	s := dbx.NewScope(conn, &testRow{})
	d := mock.New()
	d.FindErr = errors.New("find boom")
	bad := dbx.NewScope(d.Conn(), &testRow{})
	var rows []testRow
	if _, err := bad.PaginationQueryWithOpt(&rows, nil); err == nil {
		t.Fatal("pagination find error")
	}
	if _, err := bad.QueryPagination(&rows, 1, 1, false); err == nil {
		t.Fatal("query pagination find error")
	}
	if _, err := bad.ScrollQuery(&rows, "", 1); err == nil {
		t.Fatal("scroll find error")
	}
	_ = s
}
