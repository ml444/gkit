package sqlx

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ml444/gkit/dbx"
)

func TestTxPathsAndCompileErrors(t *testing.T) {
	raw, err := sqlx.Connect("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = raw.Close() })
	if _, err := raw.Exec(`CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT, score INTEGER)`); err != nil {
		t.Fatal(err)
	}
	conn := NewConn(raw).(*Conn)
	d := &Driver{db: raw} // nil ctx -> context.Background()
	_ = d.context()

	empty := &dbx.QueryBuilder{}
	if err := d.Find(context.Background(), empty, &[]struct{}{}); err == nil {
		t.Fatal("find compile")
	}
	if err := d.First(context.Background(), empty, &struct{}{}); err == nil {
		t.Fatal("first compile")
	}
	if _, err := d.Count(context.Background(), empty); err == nil {
		t.Fatal("count compile")
	}
	if _, err := d.Create(context.Background(), empty, map[string]any{"id": 1}); err == nil {
		t.Fatal("create compile")
	}
	if _, err := d.Create(context.Background(), &dbx.QueryBuilder{Table: "items"}, 123); err == nil {
		t.Fatal("create structColumns")
	}
	if _, err := d.Update(context.Background(), empty, map[string]any{"name": "x"}); err == nil {
		t.Fatal("update compile")
	}
	if _, err := d.Update(context.Background(), &dbx.QueryBuilder{Table: "items"}, 123); err == nil {
		t.Fatal("update structColumns")
	}
	if _, err := d.UpdateColumn(context.Background(), empty, "name", "x"); err == nil {
		t.Fatal("update column compile")
	}
	if _, err := d.Delete(context.Background(), empty); err == nil {
		t.Fatal("delete compile")
	}

	batch := &[]map[string]any{{"id": int64(1), "name": "a", "score": int64(1)}}
	if _, err := d.CreateInBatches(context.Background(), &dbx.QueryBuilder{Table: "items"}, batch, 1); err != nil {
		t.Fatal(err)
	}

	if err := d.Transaction(context.Background(), func(tx dbx.Driver) error {
		td := tx.(*Driver)
		var rows []struct {
			ID   int64  `db:"id"`
			Name string `db:"name"`
		}
		if err := td.Find(context.Background(), &dbx.QueryBuilder{Table: "items", Selects: []string{"id", "name"}}, &rows); err != nil {
			return err
		}
		var one struct {
			ID   int64  `db:"id"`
			Name string `db:"name"`
		}
		if err := td.First(context.Background(), &dbx.QueryBuilder{Table: "items", Selects: []string{"id", "name"}}, &one); err != nil {
			return err
		}
		if _, err := td.Count(context.Background(), &dbx.QueryBuilder{Table: "items"}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	_ = conn
}
