package mock

import (
	"context"
	"testing"

	"github.com/ml444/gkit/dbx"
)

type softRow struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	DeletedAt int64  `json:"deleted_at"`
	private   int
}

func (softRow) TableName() string { return "soft" }

func TestSoftDeleteMissingAndUnexported(t *testing.T) {
	m := New()
	m.Seed("soft", map[string]any{"id": int64(1), "name": "a"}) // no deleted_at
	b := &dbx.QueryBuilder{Table: "soft", Wheres: []dbx.WhereClause{{Query: "deleted_at = 0"}}}
	var rows []softRow
	if err := m.Find(context.Background(), b, &rows); err != nil || len(rows) != 1 {
		t.Fatalf("%#v %v", rows, err)
	}
	if _, err := structToMap(&softRow{ID: 1, Name: "x", private: 9}); err != nil {
		t.Fatal(err)
	}
	partial := map[string]any{"name": "only"}
	out := &softRow{}
	if err := mapToStruct(partial, out); err != nil || out.Name != "only" {
		t.Fatalf("%#v %v", out, err)
	}
	_ = mapToStruct(map[string]any{"private": 1}, &softRow{})

	m2 := New()
	m2.Seed("items", map[string]any{"id": "bad", "name": "z"})
	var bad []row
	if err := m2.Find(context.Background(), &dbx.QueryBuilder{Table: "items"}, &bad); err == nil {
		t.Fatal("expected assignRows convert error")
	}
}

func TestCreateInBatchesPointerSlice(t *testing.T) {
	m := New()
	b := &dbx.QueryBuilder{Table: "items"}
	rows := &[]row{{ID: 1, Name: "a"}, {ID: 2, Name: "b"}}
	n, err := m.CreateInBatches(context.Background(), b, rows, 1)
	if err != nil || n != 2 {
		t.Fatalf("%d %v", n, err)
	}
}
