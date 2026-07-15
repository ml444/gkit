// Package driverstest contains database-driver conformance tests for dbx.
// It deliberately depends only on the dbx core, so adapter modules can share it.
package driverstest

import (
	"context"
	"fmt"
	"testing"

	"github.com/ml444/gkit/dbx"
)

// Model is the common table used by the driver conformance suite.
type Model struct {
	ID   int64  `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
	Age  int64  `json:"age"`
}

func (Model) TableName() string { return "dbx_contract_model" }

// Kind is a second table used to verify multi-model transactions.
type Kind struct {
	ID   int64  `json:"id" gorm:"primaryKey"`
	Kind string `json:"kind"`
}

func (Kind) TableName() string { return "dbx_contract_kind" }

// MigrationModels returns every model an adapter fixture must initialise.
func MigrationModels() []any { return []any{&Model{}, &Kind{}} }

// Factory creates a new, empty database connection for one conformance subtest.
// It must initialise the tables returned by MigrationModels.
type Factory func(t *testing.T) dbx.Conn

// Run executes the driver-agnostic dbx integration contract.
func Run(t *testing.T, newConn Factory) {
	t.Helper()

	t.Run("repository CRUD, pagination, and scroll", func(t *testing.T) {
		conn := newConn(t)
		repo := dbx.NewT[Model](func() dbx.Conn { return conn })
		ctx := context.Background()
		for _, m := range []*Model{{ID: 1, Name: "alice", Age: 10}, {ID: 2, Name: "bob", Age: 20}, {ID: 3, Name: "cathy", Age: 30}} {
			if err := repo.Create(ctx, m); err != nil {
				t.Fatalf("create %q: %v", m.Name, err)
			}
		}
		if err := repo.BatchCreate(ctx, []*Model{{ID: 4, Name: "dave", Age: 40}, {ID: 5, Name: "erin", Age: 50}}); err != nil {
			t.Fatalf("batch create: %v", err)
		}
		count, err := repo.Count(ctx)
		if err != nil || count != 5 {
			t.Fatalf("count = %d, %v; want 5, nil", count, err)
		}
		var got Model
		if err := repo.GetOneByWhere(ctx, &got, "id = ?", int64(2)); err != nil || got.Name != "bob" {
			t.Fatalf("get one = %+v, %v", got, err)
		}
		if rows, err := repo.Update(ctx, map[string]any{"age": int64(21)}, "id = ?", int64(2)); err != nil || rows != 1 {
			t.Fatalf("update = %d, %v", rows, err)
		}

		var page []*Model
		p, err := dbx.NewScope(conn, &Model{}).Order("id ASC").PaginationQuery(&page, 2, 2)
		if err != nil || p.Total != 5 || len(page) != 2 || page[0].ID != 3 {
			t.Fatalf("pagination = %+v, rows=%+v, err=%v", p, page, err)
		}
		var scrollRows []*Model
		scroll, err := dbx.NewScope(conn, &Model{}).ScrollQuery(&scrollRows, "2", 2)
		if err != nil || len(scrollRows) != 2 || scrollRows[0].ID != 3 || scroll.Cursor != "4" {
			t.Fatalf("scroll = %+v, rows=%+v, err=%v", scroll, scrollRows, err)
		}
		if err := repo.DeleteByWhere(ctx, "id = ?", int64(5)); err != nil {
			t.Fatalf("delete: %v", err)
		}
		if err := repo.GetOneByWhere(ctx, &Model{}, "id = ?", int64(5)); err == nil {
			t.Fatal("deleted row was found")
		}
	})

	t.Run("scope increment", func(t *testing.T) {
		conn := newConn(t)
		s := dbx.NewScope(conn, &Model{})
		if err := s.Create(&Model{ID: 1, Name: "alice", Age: 10}); err != nil {
			t.Fatal(err)
		}
		if err := s.Eq("id", int64(1)).UpdateColumnWithIncr("age", 2); err != nil {
			t.Fatal(err)
		}
		var got Model
		if err := s.Eq("id", int64(1)).First(&got); err != nil || got.Age != 12 {
			t.Fatalf("increment = %+v, %v", got, err)
		}
	})

	t.Run("transactions", func(t *testing.T) {
		conn := newConn(t)
		if err := dbx.TxGo(context.Background(), conn, func(d dbx.Driver) error {
			return dbx.NewScope(dbx.StaticConn(d), &Model{}).Create(&Model{ID: 1, Name: "tx", Age: 1})
		}); err != nil {
			t.Fatalf("TxGo: %v", err)
		}
		if err := dbx.TxGo(context.Background(), conn, func(dbx.Driver) error { panic("expected") }); err == nil {
			t.Fatal("panic transaction returned nil")
		}
		if err := dbx.TxCreateMultiModels(context.Background(), conn, &Model{ID: 2, Name: "multi", Age: 2}, &Kind{ID: 1, Kind: "kind"}); err != nil {
			t.Fatalf("TxCreateMultiModels: %v", err)
		}
		var kind Kind
		if err := dbx.NewScope(conn, &Kind{}).Eq("id", int64(1)).First(&kind); err != nil || kind.Kind != "kind" {
			t.Fatalf("transaction result = %+v, %v", kind, err)
		}
	})

	t.Run("repository encryption", func(t *testing.T) {
		conn := newConn(t)
		repo := dbx.NewT[Model](func() dbx.Conn { return conn }, dbx.SetSpecifyFieldCipherMap(map[string]dbx.FieldCipher{"name": {"Name", prefixCipher{}}}))
		value := map[string]any{"name": "alice"}
		if err := repo.CheckAndCrypto(value, dbx.CipherKindEncrypt, false); err != nil || value["name"] != "enc:alice" {
			t.Fatalf("encrypt = %#v, %v", value, err)
		}
		if err := repo.CheckAndCrypto(value, dbx.CipherKindDecrypt, false); err != nil || value["name"] != "alice" {
			t.Fatalf("decrypt = %#v, %v", value, err)
		}
	})
}

type prefixCipher struct{}

func (prefixCipher) Encrypt(v any) (any, error) { return "enc:" + fmt.Sprint(v), nil }
func (prefixCipher) Decrypt(v any) (any, error) {
	s, ok := v.(string)
	if !ok || len(s) < 4 || s[:4] != "enc:" {
		return nil, fmt.Errorf("invalid encrypted value %v", v)
	}
	return s[4:], nil
}
