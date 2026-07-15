package pagination

import "testing"

func TestNewDefaultPagination(t *testing.T) {
	p := NewDefaultPagination()
	if p.Page != 1 || p.Size != 10 || p.Total != 0 || p.SkipCount {
		t.Errorf("NewDefaultPagination() = %+v, want default values", p)
	}
}

func TestSetPage(t *testing.T) {
	p := NewDefaultPagination()
	p.SetPage(5)
	if p.Page != 5 {
		t.Errorf("SetPage(5) = %d, want 5", p.Page)
	}

	p.SetPage(0)
	if p.Page != 1 {
		t.Errorf("SetPage(0) = %d, want 1", p.Page)
	}
}

func TestSetNextPage(t *testing.T) {
	p := NewDefaultPagination()
	p.Total = 100
	p.Size = 10

	ok := p.SetNextPage()
	if !ok || p.Page != 2 {
		t.Errorf("SetNextPage() = %t, %d, want true, 2", ok, p.Page)
	}

	for i := 0; i < 8; i++ {
		p.SetNextPage()
	}
	ok = p.SetNextPage()
	if ok || p.Page != 11 {
		t.Errorf("SetNextPage() = %t, %d, want false, 10", ok, p.Page)
	}
}

func TestPaginationMutatorsAndOffset(t *testing.T) {
	p := NewDefaultPagination()
	if p.SetPage(0) != p || p.Page != 1 {
		t.Fatal("SetPage should normalize zero and chain")
	}
	if p.SetPageAndSize(3, 20) != p || p.Page != 3 || p.Size != 20 {
		t.Fatalf("SetPageAndSize = %+v", p)
	}
	if p.SetSize(30) != p || p.Size != 30 {
		t.Fatalf("SetSize = %+v", p)
	}
	p.SetSize(20)
	if p.SetSkipCount() != p || !p.SkipCount {
		t.Fatal("SetSkipCount should chain and set flag")
	}
	if p.Offset() != 40 {
		t.Errorf("Offset() = %d, want 40", p.Offset())
	}
	p.SetPage(1)
	if p.Offset() != 0 {
		t.Errorf("Offset page 1 = %d", p.Offset())
	}
}

func TestSetNextPageNormalizesZeroAndDetectsLastPage(t *testing.T) {
	p := &Pagination{Page: 0, Size: 2, Total: 10}
	if !p.SetNextPage() || p.Page != 2 {
		t.Fatalf("SetNextPage zero = page %d", p.Page)
	}
	p = &Pagination{Page: 4, Size: 2, Total: 10}
	if p.SetNextPage() || p.Page != 5 {
		t.Fatalf("SetNextPage last = page %d", p.Page)
	}
}

func TestScrollMutators(t *testing.T) {
	s := (&Scroll{}).SetSize(30).SetCursor("cursor-1")
	if s.Size != 30 || s.Cursor != "cursor-1" {
		t.Fatalf("Scroll = %+v", s)
	}
}
