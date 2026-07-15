package dbx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ml444/gkit/dbx"
)

func seededScope() (*dbx.Scope, dbx.Conn) {
	d, conn := testConn()
	d.Seed("test_rows", map[string]any{"id": int64(1), "name": "alice", "age": int64(10), "deleted_at": uint32(0)}, map[string]any{"id": int64(2), "name": "bob", "age": int64(20), "deleted_at": uint32(0)})
	return dbx.NewScope(conn, &testRow{}), conn
}

func TestScopeCRUDAndReads(t *testing.T) {
	s, _ := seededScope()
	if dbx.NewScopeOfPure(s.Conn(), &testRow{}).Builder().Table != "test_rows" {
		t.Fatal("pure scope table missing")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("nil model did not panic")
			}
		}()
		dbx.NewScope(s.Conn(), nil)
	}()
	if err := s.Create(&testRow{ID: 3, Name: "c"}); err != nil || s.RowsAffected != 1 {
		t.Fatalf("create: %v, %d", err, s.RowsAffected)
	}
	if err := s.Save(&testRow{ID: 4, Name: "d"}); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateInBatches([]testRow{{ID: 5}, {ID: 6}}, 1); err != nil {
		t.Fatal(err)
	}
	if err := s.Update(map[string]any{"name": "new"}, "id = ?", int64(1)); err != nil {
		t.Fatal(err)
	}
	if err := s.Update(testRow{Name: "struct"}, map[string]any{"id": int64(1)}); err != nil {
		t.Fatal(err)
	}
	if err := s.Eq("id", int64(1)).UpdateColumn("name", "changed"); err != nil {
		t.Fatal(err)
	}
	if err := s.Eq("id", int64(1)).UpdateColumnWithIncr("age", 2); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete("id = ?", int64(1)); err != nil {
		t.Fatal(err)
	}
	var one testRow
	if err := s.First(&one, "id = ?", int64(1)); err != nil || one.Name != "changed" {
		t.Fatalf("first: %#v, %v", one, err)
	}
	// Update mutates the receiver Scope; use a fresh Scope for subsequent reads.
	fresh := dbx.NewScope(s.Conn(), &testRow{})
	var rows []testRow
	if err := fresh.Find(&rows); err != nil || len(rows) < 2 {
		t.Fatalf("find: %d, %v", len(rows), err)
	}
	if err := fresh.Scan(&one, "id = ?", int64(2)); err != nil || one.ID != 2 {
		t.Fatalf("scan: %#v, %v", one, err)
	}
	exists, err := fresh.Exist("id = ?", int64(2))
	if err != nil || !exists {
		t.Fatalf("exist: %v, %v", exists, err)
	}
	if count, err := fresh.Count(); err != nil || count < 2 {
		t.Fatalf("count: %d, %v", count, err)
	}
	if err := fresh.Eq("id", int64(99)).First(&one); !errors.Is(err, dbx.ErrRecordNotFound) {
		t.Fatalf("missing first: %v", err)
	}
	if err := fresh.SetNotFoundErr(77).Eq("id", int64(99)).First(&one); !dbx.IsNotFoundErr(err, 77) {
		t.Fatalf("custom missing: %v", err)
	}
	if err := fresh.IgnoreNotFoundErr().Eq("id", int64(99)).First(&one); err != nil {
		t.Fatalf("ignored missing: %v", err)
	}
}

func TestScopePredicatesAndConfiguration(t *testing.T) {
	s, conn := seededScope()
	q := s.Eq("a", 1).Ne("b", 2).In("c", []int{1}).NotIn("d", []int{2}).Like("e", "x").LikePrefix("f", "x").LikeSuffix("g", "x").NotLike("h", "x").IsNull("i").IsNotNull("j").Between("k", 1, 2).NotBetween("l", 1, 2)
	if len(q.Builder().Wheres) != 12 || len(s.Builder().Wheres) != 0 {
		t.Fatalf("where count or fork isolation incorrect: %d/%d", len(q.Builder().Wheres), len(s.Builder().Wheres))
	}
	want := []string{"a = ?", "b != ?", "c IN ?", "d NOT IN ?", "e LIKE ?", "f LIKE ?", "g LIKE ?", "h NOT LIKE ?", "i IS NULL", "j IS NOT NULL", "k BETWEEN ? AND ?", "l NOT BETWEEN ? AND ?"}
	for i, clause := range q.Builder().Wheres {
		if clause.Query != want[i] {
			t.Fatalf("where[%d] = %q, want %q", i, clause.Query, want[i])
		}
	}
	if s.In("x", []int{}).Builder().Wheres != nil || s.NotIn("x", nil).Builder().Wheres != nil {
		t.Fatal("empty IN should no-op")
	}
	if got := s.Gt("a", 1).Gte("b", 2).Lt("c", 3).Lte("d", 4).Builder().Wheres; len(got) != 4 {
		t.Fatalf("comparison predicates = %#v", got)
	}
	q = q.Where(map[string]any{"m": 1}).Where(map[string]string{"n": "v"}).Where("o = ?", 3).Where(map[int]string{1: "x"}).Or("p = ?", 4)
	if len(q.Builder().OrWheres) != 1 || len(q.Builder().Wheres) < 16 {
		t.Fatal("where/or helpers missing")
	}
	q = q.MultiOr([]dbx.WhereClause{{Query: "x", Args: []any{1}}, {Query: "y = ?", Args: []any{2}}}).MultiOrLike([][2]string{{"name", "a"}}, true)
	if len(q.Builder().Wheres) < 18 {
		t.Fatal("multi-or missing")
	}
	opts := &dbx.QueryOpts{Selects: []string{"id"}, Where: map[string]any{"id": 1}, Between: map[string][2]any{"age": {1, 2}}, Like: map[string]string{"name": "a"}, Or: []dbx.WhereClause{{Query: "id = ?", Args: []any{1}}}, OrLike: [][2]string{{"name", "a"}}, OrBetween: map[string][2]any{"age": {1, 2}}, IsLikePrefix: true, IsOrLikePrefix: true, GroupBys: []string{"name"}, OrderBys: []dbx.OrderColumn{{Field: "id"}}, OrderBy: "name desc", Exp: []any{"age > ?", 1}}
	q = s.Query(opts)
	if s.Query(nil) != s || len(q.Builder().Selects) != 1 || len(q.Builder().Groups) != 1 {
		t.Fatal("query options missing")
	}
	fresh := dbx.NewScope(conn, &testRow{}).Order("id desc").Orders(dbx.OrderColumn{Field: "name", Desc: true}).Group("id").Groups("age").Having("count(*) > ?", 1).Joins("join x on x.id = test_rows.id").Omit("age").Select("name").Limit(2).Offset(1).ReturnColumns("id").SetForUpdate().SetIncludeDeleted().WithContext(context.WithValue(context.Background(), "x", 1))
	b := fresh.Builder()
	if len(b.OrderRaw) != 1 || len(b.Orders) != 1 || b.Having == nil || len(b.Joins) != 1 || len(b.Omits) != 1 || b.Limit != 2 || b.Offset != 1 || !b.ForUpdate || len(b.ReturningColumns) != 1 {
		t.Fatalf("builder configuration missing: %#v", b)
	}
	q = fresh
	if q.Context() == nil || q.Conn() != conn || q.Model() == nil {
		t.Fatal("scope accessors incorrect")
	}
	if err := q.Transaction(func(dbx.Driver) error { return nil }); err != nil {
		t.Fatal(err)
	}
}

func TestScopeNoOpIncrementAndEmptyExist(t *testing.T) {
	s, _ := seededScope()
	if err := s.UpdateColumnWithIncr("age", 0); err != nil {
		t.Fatalf("zero increment = %v", err)
	}
	exists, err := s.Exist()
	if err != nil || !exists {
		t.Fatalf("empty Exist = %v, %v", exists, err)
	}
}
