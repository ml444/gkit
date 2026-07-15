package sqlx_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	dbx "github.com/ml444/gkit/dbx"
	dbxmock "github.com/ml444/gkit/dbx/mock"
	sqlxdriver "github.com/ml444/gkit/dbx/sqlx"
)

type row struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Score int64  `json:"score"`
}

func (row) TableName() string { return "items" }

func TestSqlxDriverCRUD(t *testing.T) {
	raw, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(`CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT, score INTEGER)`); err != nil {
		t.Fatal(err)
	}
	conn := sqlxdriver.NewConn(raw)
	if err := dbx.NewScope(conn, &row{}).Create(&row{Name: "a"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	var got row
	if err := dbx.NewScope(conn, &row{}).Eq("name", "a").First(&got); err != nil {
		t.Fatalf("first: %v", err)
	}
	if got.Name != "a" {
		t.Fatalf("got %+v", got)
	}
}

func TestSqlxDriverExtendedOperations(t *testing.T) {
	raw, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })
	if _, err := raw.Exec(`CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT, score INTEGER)`); err != nil {
		t.Fatal(err)
	}
	conn := sqlxdriver.NewConn(raw)
	s := dbx.NewScope(conn, &row{})
	if err := s.CreateInBatches([]row{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.Eq("id", 1).Update(map[string]any{"name": "updated"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Eq("id", 1).UpdateColumn("score", 5); err != nil {
		t.Fatal(err)
	}
	if err := s.Eq("id", 1).UpdateColumnWithIncr("score", 2); err != nil {
		t.Fatal(err)
	}
	var got row
	if err := s.Eq("id", 1).Scan(&got); err != nil || got.Name != "updated" {
		t.Fatalf("Scan = %#v, %v", got, err)
	}
	if n, err := s.Count(); err != nil || n != 2 {
		t.Fatalf("Count = %d, %v", n, err)
	}
	if err := s.Save(&row{ID: 3, Name: "saved"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Eq("id", 2).Delete(); err != nil {
		t.Fatal(err)
	}
	if err := s.Eq("id", 999).First(&got); !errors.Is(err, dbx.ErrRecordNotFound) {
		t.Fatalf("First missing = %v", err)
	}
	if err := s.Transaction(func(d dbx.Driver) error {
		_, err := d.Create(context.Background(), &dbx.QueryBuilder{Table: "items"}, &row{ID: 4, Name: "transaction"})
		return err
	}, dbx.WithTxOptions(sql.TxOptions{})); err != nil {
		t.Fatalf("Transaction success: %v", err)
	}
	want := errors.New("rollback")
	if err := s.Transaction(func(dbx.Driver) error { return want }); !errors.Is(err, want) {
		t.Fatalf("Transaction error = %v", err)
	}
	manager, ok := conn.(dbx.TxManager)
	if !ok {
		t.Fatal("Conn is not TxManager")
	}
	tx, err := manager.Begin(context.Background(), dbx.WithTxOptions(sql.TxOptions{}))
	if err != nil || manager.Commit(tx) != nil {
		t.Fatalf("Begin/Commit = %v", err)
	}
	tx, err = manager.Begin(context.Background())
	if err != nil || manager.Rollback(tx) != nil {
		t.Fatalf("Begin/Rollback = %v", err)
	}
	if manager.Commit(dbxmock.New()) != sql.ErrTxDone || manager.Rollback(dbxmock.New()) != sql.ErrTxDone {
		t.Fatal("wrong driver should return ErrTxDone")
	}
	if err := dbx.NewScope(conn, &row{}).WithContext(context.Background()).Create(&row{ID: 5, Name: "context"}); err != nil {
		t.Fatalf("WithContext/Create = %v", err)
	}
}
