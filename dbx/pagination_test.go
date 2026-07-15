package dbx_test

import (
	"testing"

	"github.com/ml444/gkit/dbx"
	"github.com/ml444/gkit/dbx/pagination"
)

func TestScopePaginationQueries(t *testing.T) {
	s, _ := seededScope()
	var rows []testRow
	p, err := s.PaginationQuery(&rows, 1, 1)
	if err != nil || p.Total != 2 || len(rows) != 1 {
		t.Fatalf("pagination: %#v %#v %v", p, rows, err)
	}
	rows = nil
	p, err = s.PaginationQueryWithOpt(&rows, nil)
	if err != nil || p.Page != 1 || p.Size != 10 {
		t.Fatalf("nil opt: %#v %v", p, err)
	}
	rows = nil
	p, err = s.PaginationQueryWithOpt(&rows, &pagination.Pagination{Page: 1, Size: 1, SkipCount: true})
	if err != nil || p.Total != 0 || !p.SkipCount {
		t.Fatalf("skip count: %#v %v", p, err)
	}
	if total, err := s.QueryPagination(&rows, 1, 1, false); err != nil || total != 2 {
		t.Fatalf("query pagination: %d %v", total, err)
	}
}

func TestScopeScrollQuery(t *testing.T) {
	s, _ := seededScope()
	var rows []testRow
	scroll, err := s.ScrollQuery(&rows, "", 1)
	if err != nil || scroll.Cursor != "1" {
		t.Fatalf("empty cursor: %#v %v", scroll, err)
	}
	rows = nil
	scroll, err = s.ScrollQuery(&rows, "1", 1)
	if err != nil || scroll.Cursor != "1" {
		t.Fatalf("numeric cursor: %#v %v", scroll, err)
	}
	rows = nil
	scroll, err = s.ScrollQuery(&rows, "not-a-number", 1, "name")
	if err != nil || scroll.Cursor == "" {
		t.Fatalf("string cursor/custom field: %#v %v", scroll, err)
	}
	if _, err := (*dbx.Scope)(nil).PaginationQuery(&rows, 1, 1); err == nil {
		t.Fatal("nil scope pagination should fail")
	}
	if _, err := (*dbx.Scope)(nil).ScrollQuery(&rows, "", 1); err == nil {
		t.Fatal("nil scope scroll should fail")
	}
}
