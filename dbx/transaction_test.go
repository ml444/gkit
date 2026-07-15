package dbx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ml444/gkit/dbx"
)

type txItem struct{ preload, execute error }

func (i txItem) Preload(dbx.Driver) error { return i.preload }
func (i txItem) Execute(dbx.Driver) error { return i.execute }

type scopeTxItem struct{ preload, execute error }

func (i scopeTxItem) Preload(*dbx.T, dbx.Driver) error { return i.preload }
func (i scopeTxItem) Execute(*dbx.T, dbx.Driver) error { return i.execute }

func TestTransactionHelpers(t *testing.T) {
	d, conn := testConn()
	ctx := context.Background()
	called := false
	if err := dbx.TxGo(ctx, conn, func(dbx.Driver) error { called = true; return nil }); err != nil || !called {
		t.Fatalf("TxGo success: %v", err)
	}
	want := errors.New("stop")
	if err := dbx.TxGo(ctx, conn, func(dbx.Driver) error { return want }); !errors.Is(err, want) {
		t.Fatalf("TxGo error: %v", err)
	}
	if err := dbx.TxGo(ctx, conn, func(dbx.Driver) error { panic("boom") }); err == nil {
		t.Fatal("TxGo panic was not recovered")
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
	d, conn := testConn()
	repo := dbx.NewT[testRow](func() dbx.Conn { return conn })
	row := &testRow{ID: 1, Name: "a"}
	if err := dbx.NewInsertItem(row).Execute(d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.UpdateItem{Model: row, Where: map[string]any{"id": int64(1)}, Updates: map[string]any{"name": "b"}}).Preload(d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.UpdateItem{Model: row, Where: map[string]any{"id": int64(1)}}).Execute(d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.SaveItem{Model: row, Where: map[string]any{"id": int64(1)}}).Preload(d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.SaveItem{Model: row}).Execute(d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.ScopeInsertItem{Models: &testRow{ID: 2}}).Execute(repo, d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.ScopeUpdateItem{Model: row, Where: map[string]any{"id": int64(1)}, Updates: map[string]any{"name": "c"}}).Preload(repo, d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.ScopeUpdateItem{Model: row, Where: map[string]any{"id": int64(1)}}).Execute(repo, d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.ScopeSaveItem{Model: row, Where: map[string]any{"id": int64(1)}}).Preload(repo, d); err != nil {
		t.Fatal(err)
	}
	if err := (&dbx.ScopeSaveItem{Model: row}).Execute(repo, d); err != nil {
		t.Fatal(err)
	}
	if err := dbx.RunTxItems(ctx(), conn, dbx.NewInsertItem(&testRow{ID: 3})); err != nil {
		t.Fatal(err)
	}
}

func ctx() context.Context { return context.Background() }
