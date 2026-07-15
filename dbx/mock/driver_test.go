package mock

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/ml444/gkit/dbx"
)

type row struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (row) TableName() string { return "items" }

func TestMockDriverSeedAndCount(t *testing.T) {
	m := New()
	m.Seed("items", map[string]any{"id": 1, "name": "x"})
	s := dbx.NewScope(m.Conn(), "items")
	n, err := s.Eq("id", 1).Count()
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("count = %d", n)
	}
}

func TestMockDriverRepository(t *testing.T) {
	m := New()
	m.Seed("items", map[string]any{"id": 1, "name": "alice"})
	repo := dbx.NewT[row](func() dbx.Conn { return m.Conn() })
	var got row
	if err := repo.GetOne(context.Background(), &got, 1); err != nil {
		t.Fatalf("GetOne: %v", err)
	}
	if got.Name != "alice" {
		t.Fatalf("name = %q", got.Name)
	}
}

func TestDriverOperationsAndErrors(t *testing.T) {
	ctx := context.Background()
	m := New()
	m.Seed("items",
		map[string]any{"id": int64(1), "name": "alice", "score": int64(2)},
		map[string]any{"id": int64(2), "name": "bob", "score": int64(4)},
	)
	b := &dbx.QueryBuilder{Model: row{}}

	var values []row
	if err := m.Find(ctx, b, &values); err != nil || len(values) != 2 {
		t.Fatalf("Find values = %#v, %v", values, err)
	}
	var pointers []*row
	if err := m.Find(ctx, b, &pointers); err != nil || pointers[0].Name != "alice" {
		t.Fatalf("Find pointers = %#v, %v", pointers, err)
	}
	var got row
	if err := m.First(ctx, &dbx.QueryBuilder{Table: "items", Wheres: []dbx.WhereClause{{Query: "id = ?", Args: []any{int64(2)}}}}, &got); err != nil || got.Name != "bob" {
		t.Fatalf("First = %#v, %v", got, err)
	}
	if err := m.First(ctx, &dbx.QueryBuilder{Table: "items", Wheres: []dbx.WhereClause{{Query: "id = ?", Args: []any{int64(9)}}}}, &got); !errors.Is(err, dbx.ErrRecordNotFound) {
		t.Fatalf("First missing = %v", err)
	}
	if err := m.Scan(ctx, b, &got); err != nil || got.Name != "alice" {
		t.Fatalf("Scan = %#v, %v", got, err)
	}
	if n, err := m.Count(ctx, b); err != nil || n != 2 {
		t.Fatalf("Count = %d, %v", n, err)
	}

	if n, err := m.Create(ctx, b, row{ID: 3, Name: "cara"}); err != nil || n != 1 {
		t.Fatalf("Create = %d, %v", n, err)
	}
	if n, err := m.CreateInBatches(ctx, b, []row{{ID: 4}, {ID: 5}}, 1); err != nil || n != 2 {
		t.Fatalf("CreateInBatches = %d, %v", n, err)
	}
	if n, err := m.Save(ctx, b, row{ID: 6}); err != nil || n != 1 {
		t.Fatalf("Save = %d, %v", n, err)
	}
	if n, err := m.Update(ctx, b, map[string]any{"name": "updated"}); err != nil || n != 1 {
		t.Fatalf("Update = %d, %v", n, err)
	}
	if n, err := m.UpdateColumn(ctx, &dbx.QueryBuilder{Table: "items", Wheres: []dbx.WhereClause{{Query: "id = ?", Args: []any{int64(1)}}}}, "name", "ann"); err != nil || n != 1 {
		t.Fatalf("UpdateColumn = %d, %v", n, err)
	}
	if n, err := m.UpdateColumn(ctx, &dbx.QueryBuilder{Table: "items", Wheres: []dbx.WhereClause{{Query: "id = ?", Args: []any{int64(1)}}}, IncrColumn: "score", IncrValue: -1}, "ignored", nil); err != nil || n != 1 || m.Tables["items"][0]["score"] != int64(1) {
		t.Fatalf("increment = %d, %v, %#v", n, err, m.Tables["items"][0])
	}
	if n, err := m.Delete(ctx, b); err != nil || n != 1 {
		t.Fatalf("Delete = %d, %v", n, err)
	}
	if !reflect.DeepEqual(m.WithContext(ctx), m) {
		t.Fatal("WithContext did not return driver")
	}
	for _, injected := range []struct {
		name string
		set  func(error)
		call func() error
	}{
		{"create", func(e error) { m.CreateErr = e }, func() error { _, e := m.Create(ctx, b, row{}); return e }},
		{"find", func(e error) { m.FindErr = e }, func() error { return m.Find(ctx, b, &values) }},
		{"tx", func(e error) { m.TxErr = e }, func() error { return m.Transaction(ctx, func(dbx.Driver) error { return nil }) }},
	} {
		want := errors.New(injected.name)
		injected.set(want)
		if err := injected.call(); !errors.Is(err, want) {
			t.Fatalf("%s injected error = %v", injected.name, err)
		}
	}
}

func TestDriverHelpersAndConnection(t *testing.T) {
	m := New()
	conn := m.Conn()
	if conn.Driver(context.Background()) != m {
		t.Fatal("Conn.Driver mismatch")
	}
	d, err := conn.(*Conn).Begin(context.Background())
	if err != nil || d != m || conn.(*Conn).Commit(d) != nil || conn.(*Conn).Rollback(d) != nil {
		t.Fatalf("transaction manager = %v, %v", d, err)
	}
	if err := m.Transaction(context.Background(), func(d dbx.Driver) error { return errors.New("stop") }); err == nil {
		t.Fatal("Transaction should return callback error")
	}
	if _, err := m.Create(context.Background(), &dbx.QueryBuilder{Table: "items"}, 1); err == nil {
		t.Fatal("Create non-struct should fail")
	}
	if err := assignRows(row{}, nil); err == nil {
		t.Fatal("assignRows non-pointer should fail")
	}
	var one row
	if err := assignRows(&one, nil); err == nil {
		t.Fatal("assignRows non-slice should fail")
	}
	tagged := struct {
		ID      int    `json:"id,omitempty"`
		Deleted uint64 `json:"deleted_at"`
	}{ID: 7}
	mapped, err := structToMap(tagged)
	if err != nil || mapped["id"] != 7 {
		t.Fatalf("structToMap = %#v, %v", mapped, err)
	}
	for _, deleted := range []any{int(0), int64(0), uint32(0), uint64(0)} {
		if !matchWhere(map[string]any{"deleted_at": deleted}, dbx.WhereClause{Query: "deleted_at = 0"}) {
			t.Fatalf("soft-delete value %T did not match", deleted)
		}
	}
	if key, ok := parseEq("name = ?"); !ok || key != "name" {
		t.Fatalf("parseEq = %q, %t", key, ok)
	}
	if _, ok := parseEq("name=?"); ok {
		t.Fatal("parseEq accepted malformed query")
	}
}

func TestDriverHelperErrorAndNumericBranches(t *testing.T) {
	for _, value := range []any{uint(1), uint64(2), uint32(3)} {
		got, ok := toInt64(value)
		if !ok || got == 0 {
			t.Fatalf("toInt64(%T) = %d, %t", value, got, ok)
		}
	}
	if _, ok := toInt64("not-a-number"); ok {
		t.Fatal("toInt64 should reject strings")
	}
	if dbxTableName(nil) != "" || dbxTableName(struct{}{}) != "" {
		t.Fatal("dbxTableName should return an empty name")
	}
	if err := assignRow(row{}, map[string]any{}); err == nil {
		t.Fatal("assignRow non-pointer should fail")
	}

	m := New()
	b := &dbx.QueryBuilder{Table: "items"}
	rows, err := m.CreateInBatches(context.Background(), b, []any{row{ID: 1}, 1}, 1)
	if err == nil || rows != 1 {
		t.Fatalf("CreateInBatches = %d, %v", rows, err)
	}
	m.FindErr = errors.New("find")
	if err := m.First(context.Background(), b, &row{}); !errors.Is(err, m.FindErr) {
		t.Fatalf("First FindErr = %v", err)
	}

	m2 := New()
	m2.Seed("items",
		map[string]any{"id": int64(1), "name": "a"},
		map[string]any{"id": int64(2), "name": "b"},
		map[string]any{"id": int64(3), "name": "c"},
	)
	if n, err := m2.Count(context.Background(), &dbx.QueryBuilder{Table: "items", Limit: 2}); err != nil || n != 2 {
		t.Fatalf("limit count = %d, %v", n, err)
	}
	if n, err := m2.UpdateColumn(context.Background(), &dbx.QueryBuilder{Table: "items", Wheres: []dbx.WhereClause{{Query: "id = ?", Args: []any{int64(99)}}}}, "name", "x"); err != nil || n != 0 {
		t.Fatalf("update missing = %d, %v", n, err)
	}
	if got, ok := toInt64(int(9)); !ok || got != 9 {
		t.Fatalf("toInt64 int = %d, %t", got, ok)
	}
	if matchWhere(map[string]any{"deleted_at": int(1)}, dbx.WhereClause{Query: "deleted_at = 0"}) {
		t.Fatal("deleted_at int 1 should not match")
	}
}
