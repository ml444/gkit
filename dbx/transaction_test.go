package dbx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ml444/gkit/dbx"
)

type txItem struct{ preload, execute error }

func (i txItem) Preload(ctx context.Context, d dbx.Driver) error { return i.preload }
func (i txItem) Execute(ctx context.Context, d dbx.Driver) error { return i.execute }

type scopeTxItem struct{ preload, execute error }

func (i scopeTxItem) Preload(ctx context.Context, t *dbx.T, d dbx.Driver) error { return i.preload }
func (i scopeTxItem) Execute(ctx context.Context, t *dbx.T, d dbx.Driver) error { return i.execute }

func TestTransactionHelpers(t *testing.T) {
	d, conn := testConn()
	ctx := context.Background()
	called := false
	if err := dbx.TxGo(ctx, conn, func(dbx.Driver) error { called = true; return nil }); err != nil || !called {
		t.Fatalf("dbx.TxGo success: %v", err)
	}
	want := errors.New("stop")
	if err := dbx.TxGo(ctx, conn, func(dbx.Driver) error { return want }); !errors.Is(err, want) {
		t.Fatalf("dbx.TxGo error: %v", err)
	}
	if err := dbx.TxGo(ctx, conn, func(dbx.Driver) error { panic("boom") }); err == nil {
		t.Fatal("dbx.TxGo panic was not recovered")
	}
	if err := dbx.ScopeTxGo(ctx, conn, func() (any, func(*dbx.Scope) error) {
		return &testRow{ID: 1}, func(s *dbx.Scope) error { return s.Create(&testRow{ID: 1}) }
	}); err != nil {
		t.Fatal(err)
	}
	repo := dbx.NewT[testRow](func() dbx.Conn { return conn })
	if err := dbx.ScopeTxGoWithT(ctx, repo, func() (any, func(*dbx.Scope) error) {
		return nil, func(s *dbx.Scope) error { return s.Create(&testRow{ID: 2}) }
	}); err != nil {
		t.Fatal(err)
	}
	if err := dbx.TxCreateMultiModels(ctx, conn, &testRow{ID: 3}, &encryptRow{ID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := dbx.RunTxItems(ctx, conn); err != nil {
		t.Fatal(err)
	}
	if err := dbx.RunTxItems(ctx, conn, txItem{}); err != nil {
		t.Fatal(err)
	}
	if err := dbx.RunTxItems(ctx, conn, txItem{preload: want}); !errors.Is(err, want) {
		t.Fatalf("preload error: %v", err)
	}
	if err := dbx.RunTxItemsWithT(ctx, repo, scopeTxItem{}); err != nil {
		t.Fatal(err)
	}
	if d == nil {
		t.Fatal("sanity")
	}
}

func TestTransactionItems(t *testing.T) {
	ctx := context.Background()
	_, conn := testConn()
	d := conn.Driver(ctx)
	repo := dbx.NewT[testRow](func() dbx.Conn { return conn })
	row := &testRow{ID: 1, Name: "a"}
	if err := dbx.NewInsertItem(row).Execute(ctx, d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.UpdateItem{Model: row, Where: map[string]any{"id": int64(1)}, Updates: map[string]any{"name": "b"}}).Preload(ctx, d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.UpdateItem{Model: row, Where: map[string]any{"id": int64(1)}}).Execute(ctx, d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.SaveItem{Model: row, Where: map[string]any{"id": int64(1)}}).Preload(ctx, d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.SaveItem{Model: row}).Execute(ctx, d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.ScopeInsertItem{Models: &testRow{ID: 2}}).Execute(ctx, repo, d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.ScopeUpdateItem{Model: row, Where: map[string]any{"id": int64(1)}, Updates: map[string]any{"name": "c"}}).Preload(ctx, repo, d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.ScopeUpdateItem{Model: row, Where: map[string]any{"id": int64(1)}}).Execute(ctx, repo, d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.ScopeSaveItem{Model: row, Where: map[string]any{"id": int64(1)}}).Preload(ctx, repo, d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.ScopeSaveItem{Model: row}).Execute(ctx, repo, d); err != nil {
		t.Fatal(err)
	}
	if err := dbx.RunTxItems(ctx, conn, dbx.NewInsertItem(&testRow{ID: 3})); err != nil {
		t.Fatal(err)
	}
}

func ctx() context.Context { return context.Background() }

type softUpdatedRow struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	UpdatedAt int64  `json:"updated_at"`
	DeletedAt int64  `json:"deleted_at"`
}

func (softUpdatedRow) TableName() string     { return "soft_updated" }
func (s softUpdatedRow) GetDeletedAt() int64 { return s.DeletedAt }
func (s softUpdatedRow) GetUpdatedAt() int64 { return s.UpdatedAt }

type plainDriver struct{}

func (plainDriver) Find(ctx context.Context, b *dbx.QueryBuilder, dest any) error { return nil }
func (plainDriver) First(ctx context.Context, b *dbx.QueryBuilder, dest any) error {
	return dbx.ErrRecordNotFound
}
func (plainDriver) Count(ctx context.Context, b *dbx.QueryBuilder) (int64, error) { return 0, nil }
func (plainDriver) Create(ctx context.Context, b *dbx.QueryBuilder, v any) (int64, error) {
	return 1, nil
}
func (plainDriver) CreateInBatches(ctx context.Context, b *dbx.QueryBuilder, values any, batchSize int) (int64, error) {
	return 1, nil
}
func (plainDriver) Save(ctx context.Context, b *dbx.QueryBuilder, v any) (int64, error) {
	return 1, nil
}
func (plainDriver) Update(ctx context.Context, b *dbx.QueryBuilder, v any) (int64, error) {
	return 0, errors.New("upd")
}
func (plainDriver) UpdateColumn(ctx context.Context, b *dbx.QueryBuilder, field string, value any) (int64, error) {
	return 0, errors.New("upc")
}
func (plainDriver) Delete(ctx context.Context, b *dbx.QueryBuilder) (int64, error) { return 0, nil }
func (plainDriver) Scan(ctx context.Context, b *dbx.QueryBuilder, dest any) error {
	return dbx.ErrRecordNotFound
}
func (plainDriver) Transaction(ctx context.Context, fn func(dbx.Driver) error, opts ...dbx.TxOption) error {
	return fn(plainDriver{})
}

type plainConn struct{ d dbx.Driver }

func (c plainConn) Driver(ctx context.Context) dbx.Driver { return c.d }

func TestProtoUpdatedAndPlainDriverTx(t *testing.T) {
	_, conn := testConn()
	s := dbx.NewScope(conn, softUpdatedRow{})
	if err := s.Update(map[string]any{"name": "n"}, "id = ?", int64(1)); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	pc := plainConn{d: plainDriver{}}
	if err := dbx.TxGo(context.Background(), pc, func(d dbx.Driver) error {
		return dbx.NewScope(dbx.StaticConn(d), &testRow{}).Create(&testRow{ID: 1})
	}); err != nil {
		t.Fatal(err)
	}
	s2 := dbx.NewScope(pc, &testRow{})
	_ = s2.Update(map[string]any{"name": "x"})
	_ = s2.UpdateColumn("name", "y")
	_ = s2.UpdateColumnWithIncr("age", 1)

	repo := dbx.NewT[srcModel](func() dbx.Conn { return conn })
	if err := repo.BatchCreate(context.Background(), &srcModel{ID: 8, Name: "solo"}); err != nil {
		t.Fatal(err)
	}
	var list []srcModel
	_ = repo.ListAll(context.Background(), &list, map[string]string{"name": "alice"})

	er := dbx.NewT[encryptRow](func() dbx.Conn { return conn }, encryptRowCrypto())
	_ = er.Create(ctx, &encryptRow{Name: "a"})
	_ = er.GetOneByWhere(ctx, &encryptRow{}, map[string]any{"name": "a"})

	d := conn.Driver(ctx)
	su := &dbx.ScopeUpdateItem{Model: &encryptRow{ID: 1}, Where: map[string]any{"name": "a"}, Updates: nil}
	_ = su.Preload(ctx, er, d)
	_ = su.Execute(ctx, er, d)
	ss := &dbx.ScopeSaveItem{Model: &encryptRow{ID: 1, Name: "a"}, Where: map[string]any{"name": "a"}}
	_ = ss.Preload(ctx, er, d)
	_ = ss.Execute(ctx, er, d)
	si := &dbx.SaveItem{Model: &testRow{ID: 1}, Where: map[string]any{"id": int64(1)}}
	_ = si.Preload(ctx, d)
	upd := &dbx.UpdateItem{Model: &testRow{}, Where: map[string]any{"id": int64(1)}, Updates: nil}
	_ = upd.Execute(ctx, d)

	er2 := dbx.NewT[storageEncryptRow](func() dbx.Conn { return conn }, dualFieldCrypto())
	_, _ = er2.ExistByWhere(context.Background(), map[string]any{"secret": "alice"})
	_ = er2.ListAll(context.Background(), &[]storageEncryptRow{}, map[string]any{"secret": "alice"})

	d3, c3 := testConn()
	fr := dbx.NewT[srcModel](func() dbx.Conn { return c3 }, dbx.WithTNotFoundErrCode(3))
	_ = fr.GetOne(context.Background(), &srcModel{}, uint64(404))
	_ = d3

	// force GetOne success already covered; trigger Find error on ListAll
	d3.FindErr = errors.New("listfail")
	_ = fr.ListAll(context.Background(), &list, nil)
	_, _ = fr.ListWithPagination(context.Background(), &list, nil, 1, 1)
}

func TestTransactionItemBranches(t *testing.T) {
	dr, conn := testConn()
	dr.Seed("test_rows", map[string]any{"id": int64(1), "name": "a", "age": int64(1), "deleted_at": uint32(0)})
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
	d := conn.Driver(ctx)
	_ = updNil.Execute(ctx, d)
	updEmpty := &dbx.UpdateItem{Model: &testRow{ID: 1}, Where: map[string]any{"id": int64(1)}}
	if err := updEmpty.Execute(ctx, d); err != nil {
		t.Fatal(err)
	}

	save := &dbx.SaveItem{Model: &testRow{ID: 10, Name: "s"}, Where: map[string]any{"id": int64(999)}}
	if err := dbx.RunTxItems(ctx, conn, save); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.SaveItem{}).Preload(ctx, d); err == nil {
		t.Fatal("SaveItem nil model")
	}
	_ = (&dbx.SaveItem{Model: &testRow{ID: 11}}).Execute(ctx, d)

	si := &dbx.ScopeInsertItem{Models: nil}
	_ = si.Preload(ctx, repo, d)
	_ = si.Execute(ctx, repo, d)
	if err := dbx.RunTxItemsWithT(ctx, repo, &dbx.ScopeInsertItem{Models: &testRow{ID: 12, Name: "si"}}); err != nil {
		t.Fatal(err)
	}

	su := &dbx.ScopeUpdateItem{Model: &testRow{}, Where: map[string]any{"id": int64(1)}, Updates: map[string]any{"name": "su"}}
	if err := dbx.RunTxItemsWithT(ctx, repo, su); err != nil {
		t.Fatal(err)
	}
	_ = (&dbx.ScopeUpdateItem{Model: nil, Where: nil}).Execute(ctx, repo, d)
	su2 := &dbx.ScopeUpdateItem{Model: &testRow{ID: 1}, Where: map[string]any{"id": int64(1)}}
	if err := su2.Execute(ctx, repo, d); err != nil {
		t.Fatal(err)
	}

	ss := &dbx.ScopeSaveItem{Model: &testRow{ID: 13, Name: "ss"}, Where: map[string]any{"id": int64(888)}}
	if err := dbx.RunTxItemsWithT(ctx, repo, ss); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.ScopeSaveItem{}).Preload(ctx, repo, d); err == nil {
		t.Fatal("ScopeSaveItem nil")
	}
	_ = (&dbx.ScopeSaveItem{Model: &testRow{ID: 14}}).Execute(ctx, repo, d)

	if err := dbx.ScopeTxGoWithT(ctx, repo, func() (any, func(*dbx.Scope) error) {
		return nil, func(s *dbx.Scope) error { return s.Create(&testRow{ID: 15, Name: "tx"}) }
	}); err != nil {
		t.Fatal(err)
	}
}