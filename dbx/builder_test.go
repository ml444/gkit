package dbx

import "testing"

type plainTestModel struct{ SomeID int }

func TestQueryBuilderConstructionAndClone(t *testing.T) {
	if b := newQueryBuilder("widgets"); b.Table != "widgets" || b.Model != nil {
		t.Fatalf("string builder: %#v", b)
	}
	if b := newQueryBuilder(testRow{}); b.Table != "test_rows" {
		t.Fatalf("table builder: %#v", b)
	}
	if b := newQueryBuilder(plainTestModel{}); b.Model == nil || b.Table != "" {
		t.Fatalf("plain builder: %#v", b)
	}
	if tableNameFromModel(nil) != "" || tableNameFromModel(&plainTestModel{}) != "plain_test_model" || tableNameFromModel(testRow{}) != "test_rows" {
		t.Fatal("table names incorrect")
	}
	b := &QueryBuilder{Selects: []string{"a"}, Omits: []string{"b"}, Wheres: []WhereClause{{Query: "x", Args: []any{1}}}, OrWheres: []WhereClause{{Query: "y"}}, Orders: []OrderColumn{{Field: "id"}}, OrderRaw: []string{"id desc"}, Groups: []string{"x"}, Having: &WhereClause{Query: "count(*) > ?", Args: []any{1}}, Joins: []joinClause{{Query: "join x"}}, ReturningColumns: []string{"id"}}
	c := b.Clone()
	c.Selects[0], c.Wheres[0].Query, c.Having.Query = "z", "changed", "changed"
	if b.Selects[0] != "a" || b.Wheres[0].Query != "x" || b.Having.Query != "count(*) > ?" {
		t.Fatal("clone modified source")
	}
	if (&QueryBuilder{}).Clone() == nil || (*QueryBuilder)(nil).Clone() == nil {
		t.Fatal("clone should never be nil")
	}
}

func TestQueryBuilderWhereHelpers(t *testing.T) {
	b := &QueryBuilder{}
	b.addMapWhere(map[string]any{"name": "a", "age": 1})
	if len(b.Wheres) != 2 {
		t.Fatalf("map where count = %d", len(b.Wheres))
	}
	b.Wheres = append(b.Wheres, WhereClause{Query: "deleted_at = 0"}, WhereClause{Query: "x = ?", Args: []any{1}}, WhereClause{Query: "deleted_at = 0", Args: []any{1}})
	b.removeSoftDeleteFilter()
	for _, w := range b.Wheres {
		if w.Query == "deleted_at = 0" && len(w.Args) == 0 {
			t.Fatal("soft delete filter remained")
		}
	}
}

func TestTableNameFromNonStruct(t *testing.T) {
	if tableNameFromModel(123) != "" {
		t.Fatal("non-struct should be empty")
	}
}

func TestNilForkAndContext(t *testing.T) {
	if (*Scope)(nil).fork() != nil {
		t.Fatal("nil fork")
	}
	s := &Scope{}
	if s.context() == nil {
		t.Fatal("nil ctx fallback")
	}
}
