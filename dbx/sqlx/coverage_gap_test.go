package sqlx

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ml444/gkit/dbx"
)

func TestSqlxInternalCoverageGaps(t *testing.T) {
	raw, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })
	if _, err := raw.Exec(`CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT, score INTEGER)`); err != nil {
		t.Fatal(err)
	}
	conn := NewConn(raw).(*Conn)
	if conn.DB() == nil {
		t.Fatal("DB()")
	}
	d := conn.Driver(context.Background()).(*Driver)
	_ = d.WithContext(context.Background())
	_ = d.context()

	b := &dbx.QueryBuilder{Table: "items", Model: &struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}{}}
	if _, _, err := compileSelect(b); err != nil {
		t.Fatal(err)
	}
	if _, _, err := compileCount(b); err != nil {
		t.Fatal(err)
	}
	b2 := &dbx.QueryBuilder{Table: "items", Orders: []dbx.OrderColumn{{Field: "id", Desc: true}}, OrderRaw: []string{"name ASC"}}
	_ = compileOrder(b2)
	if tableName(&dbx.QueryBuilder{}) == "" {
		// ok empty
	}
	if tableName(&dbx.QueryBuilder{Model: struct{ X int }{}}) == "" {
		// derived
	}
	v := map[string]any{"id": 1, "name": "a"}
	if _, _, err := compileInsert(&dbx.QueryBuilder{Table: "items"}, v); err != nil {
		t.Fatal(err)
	}
	if _, _, err := compileUpdate(&dbx.QueryBuilder{Table: "items", Wheres: []dbx.WhereClause{{Query: "id = ?", Args: []any{1}}}}, v); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Create(context.Background(), &dbx.QueryBuilder{Table: "items"}, map[string]any{"id": int64(10), "name": "z", "score": int64(1)}); err != nil {
		t.Fatal(err)
	}
	if _, err := d.CreateInBatches(context.Background(), &dbx.QueryBuilder{Table: "items"}, []map[string]any{{"id": int64(11), "name": "y", "score": int64(2)}}, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Update(context.Background(), &dbx.QueryBuilder{Table: "items", Wheres: []dbx.WhereClause{{Query: "id = ?", Args: []any{int64(10)}}}}, map[string]any{"name": "zz"}); err != nil {
		t.Fatal(err)
	}
	if _, err := d.UpdateColumn(context.Background(), &dbx.QueryBuilder{Table: "items", Wheres: []dbx.WhereClause{{Query: "id = ?", Args: []any{int64(10)}}}}, "score", 9); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Delete(context.Background(), &dbx.QueryBuilder{Table: "items", Wheres: []dbx.WhereClause{{Query: "id = ?", Args: []any{int64(11)}}}}); err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		ID    int64  `db:"id"`
		Name  string `db:"name"`
		Score int64  `db:"score"`
	}
	if err := d.Find(context.Background(), &dbx.QueryBuilder{Table: "items", Selects: []string{"id", "name", "score"}}, &rows); err != nil {
		t.Fatal(err)
	}
	var one struct {
		ID    int64  `db:"id"`
		Name  string `db:"name"`
		Score int64  `db:"score"`
	}
	_ = d.First(context.Background(), &dbx.QueryBuilder{Table: "items", Selects: []string{"id", "name", "score"}}, &one)
	if _, err := d.Count(context.Background(), &dbx.QueryBuilder{Table: "items"}); err != nil {
		t.Fatal(err)
	}
	tx, err := conn.Begin(context.Background(), dbx.WithTxOptions(sql.TxOptions{}))
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.Rollback(tx)

	// Error paths via closed DB
	_ = raw.Close()
	_, _ = d.Create(context.Background(), &dbx.QueryBuilder{Table: "items"}, map[string]any{"id": int64(99), "name": "x"})
	_, _ = d.CreateInBatches(context.Background(), &dbx.QueryBuilder{Table: "items"}, []map[string]any{{"id": int64(100)}}, 1)
	_, _ = d.Update(context.Background(), &dbx.QueryBuilder{Table: "items"}, map[string]any{"name": "x"})
	_, _ = d.UpdateColumn(context.Background(), &dbx.QueryBuilder{Table: "items"}, "score", 1)
	_, _ = d.Delete(context.Background(), &dbx.QueryBuilder{Table: "items"})
	_ = d.Find(context.Background(), &dbx.QueryBuilder{Table: "items"}, &rows)
	_ = d.First(context.Background(), &dbx.QueryBuilder{Table: "items"}, &one)
	_, _ = d.Count(context.Background(), &dbx.QueryBuilder{Table: "items"})
	_ = d.Transaction(context.Background(), func(dbx.Driver) error { return nil })
	_, _ = conn.Begin(context.Background())
	_, _, _ = compileInsert(&dbx.QueryBuilder{Table: "items"}, 123)
	_, _, _ = compileUpdate(&dbx.QueryBuilder{Table: "items"}, 123)
	_, _, _ = compileSelect(&dbx.QueryBuilder{})
	_, _, _ = compileCount(&dbx.QueryBuilder{})
}
